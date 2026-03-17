package leads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var validate = validator.New()

// Service handles lead capture operations.
type Service struct {
	db        *pgxpool.Pool
	notifier  Notifier
	log       *zap.Logger
	httpClient *http.Client
}

// Notifier is an interface for sending lead notifications.
type Notifier interface {
	NotifyLead(ctx context.Context, lead *Lead, tenantName, webhookURL, email, discordURL string) error
}

// NewService creates a new leads Service.
func NewService(db *pgxpool.Pool, notifier Notifier, log *zap.Logger) *Service {
	return &Service{
		db:         db,
		notifier:   notifier,
		log:        log,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Capture validates, deduplicates, and stores a new lead.
func (s *Service) Capture(ctx context.Context, tenantID string, req CaptureRequest) (*Lead, error) {
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// Deduplicate: check if email captured for this tenant in last 7 days
	var existing int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM leads WHERE tenant_id=$1 AND email=$2 AND created_at > NOW() - INTERVAL '7 days'`,
		tenantID, strings.ToLower(req.Email),
	).Scan(&existing)
	if err != nil {
		s.log.Error("dedup check failed", zap.Error(err))
	}
	if existing > 0 {
		// Return success but don't create duplicate
		var lead Lead
		err := s.db.QueryRow(ctx,
			`SELECT id, tenant_id, first_name, last_name, email, phone, source_url, status, created_at
			 FROM leads WHERE tenant_id=$1 AND email=$2 ORDER BY created_at DESC LIMIT 1`,
			tenantID, strings.ToLower(req.Email),
		).Scan(&lead.ID, &lead.TenantID, &lead.FirstName, &lead.LastName, &lead.Email,
			&lead.Phone, &lead.SourceURL, &lead.Status, &lead.CreatedAt)
		if err == nil {
			return &lead, nil
		}
	}

	lead := &Lead{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          strings.ToLower(req.Email),
		Phone:          req.Phone,
		SourceURL:      req.SourceURL,
		UTMSource:      req.UTMSource,
		UTMMedium:      req.UTMMedium,
		UTMCampaign:    req.UTMCampaign,
		SessionSummary: req.SessionSummary,
		Status:         "new",
		CreatedAt:      time.Now().UTC(),
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO leads (id, tenant_id, first_name, last_name, email, phone,
		                   source_url, utm_source, utm_medium, utm_campaign,
		                   session_summary, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		lead.ID, lead.TenantID, lead.FirstName, lead.LastName, lead.Email, lead.Phone,
		lead.SourceURL, lead.UTMSource, lead.UTMMedium, lead.UTMCampaign,
		lead.SessionSummary, lead.Status, lead.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert lead: %w", err)
	}

	return lead, nil
}

// DispatchWebhook sends the lead payload to the tenant's webhook URL asynchronously.
func (s *Service) DispatchWebhook(lead *Lead, webhookURL string) {
	if webhookURL == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		payload := WebhookPayload{
			Event:    "lead.captured",
			TenantID: lead.TenantID,
			Lead:     *lead,
		}

		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL,
			bytes.NewReader(body))
		if err != nil {
			s.log.Error("create webhook request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Blueprint-Event", "lead.captured")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			s.log.Error("dispatch webhook", zap.Error(err), zap.String("url", webhookURL))
			return
		}
		defer resp.Body.Close()

		// Mark webhook as sent
		_, dbErr := s.db.Exec(context.Background(),
			`UPDATE leads SET webhook_sent=true, webhook_sent_at=$2 WHERE id=$1`,
			lead.ID, time.Now().UTC())
		if dbErr != nil {
			s.log.Error("mark webhook sent", zap.Error(dbErr))
		}
	}()
}

// List returns leads for a tenant with optional filtering.
func (s *Service) List(ctx context.Context, tenantID string, limit, offset int) ([]*Lead, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, first_name, last_name, email, phone,
		       source_url, utm_source, utm_medium, utm_campaign,
		       session_summary, status, webhook_sent, created_at
		FROM leads WHERE tenant_id=$1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list leads: %w", err)
	}
	defer rows.Close()

	var leads []*Lead
	for rows.Next() {
		var l Lead
		err := rows.Scan(
			&l.ID, &l.TenantID, &l.FirstName, &l.LastName, &l.Email, &l.Phone,
			&l.SourceURL, &l.UTMSource, &l.UTMMedium, &l.UTMCampaign,
			&l.SessionSummary, &l.Status, &l.WebhookSent, &l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		leads = append(leads, &l)
	}
	return leads, rows.Err()
}

// UpdateStatus updates the CRM status of a lead.
func (s *Service) UpdateStatus(ctx context.Context, leadID, status string) error {
	_, err := s.db.Exec(ctx, `UPDATE leads SET status=$2 WHERE id=$1`, leadID, status)
	return err
}
