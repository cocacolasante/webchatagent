package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blueprintautomation/blueprint-chat/internal/knowledge"
	"github.com/blueprintautomation/blueprint-chat/pkg/encrypt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Service provides tenant CRUD operations.
type Service struct {
	db            *pgxpool.Pool
	encryptionKey string
	log           *zap.Logger
}

// NewService creates a new tenant Service.
func NewService(db *pgxpool.Pool, encryptionKey string, log *zap.Logger) *Service {
	return &Service{db: db, encryptionKey: encryptionKey, log: log}
}

// GetByID retrieves a tenant by its UUID.
func (s *Service) GetByID(ctx context.Context, id string) (*Tenant, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, api_key, name, plan, is_active,
		       bot_name, avatar_url, primary_color, greeting, position,
		       business_info, knowledge_base,
		       scheduler_type, scheduler_config,
		       lead_capture_enabled, lead_form_config, lead_webhook_url, lead_notify_email,
		       discord_webhook_url, max_messages_per_day, max_messages_per_min,
		       created_at, updated_at,
		       partner_id, managed_by, client_id
		FROM tenants WHERE id = $1`, id)

	return scanTenant(row)
}

// GetByAPIKey retrieves a tenant by its API key.
func (s *Service) GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, api_key, name, plan, is_active,
		       bot_name, avatar_url, primary_color, greeting, position,
		       business_info, knowledge_base,
		       scheduler_type, scheduler_config,
		       lead_capture_enabled, lead_form_config, lead_webhook_url, lead_notify_email,
		       discord_webhook_url, max_messages_per_day, max_messages_per_min,
		       created_at, updated_at,
		       partner_id, managed_by, client_id
		FROM tenants WHERE api_key = $1 AND is_active = true`, apiKey)

	return scanTenant(row)
}

// List returns all tenants.
func (s *Service) List(ctx context.Context) ([]*Tenant, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, api_key, name, plan, is_active,
		       bot_name, avatar_url, primary_color, greeting, position,
		       business_info, knowledge_base,
		       scheduler_type, scheduler_config,
		       lead_capture_enabled, lead_form_config, lead_webhook_url, lead_notify_email,
		       discord_webhook_url, max_messages_per_day, max_messages_per_min,
		       created_at, updated_at,
		       partner_id, managed_by, client_id
		FROM tenants ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		t, err := scanTenantRow(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// ListByPartner returns tenants belonging to a specific partner.
func (s *Service) ListByPartner(ctx context.Context, partnerID string) ([]*Tenant, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, api_key, name, plan, is_active,
		       bot_name, avatar_url, primary_color, greeting, position,
		       business_info, knowledge_base,
		       scheduler_type, scheduler_config,
		       lead_capture_enabled, lead_form_config, lead_webhook_url, lead_notify_email,
		       discord_webhook_url, max_messages_per_day, max_messages_per_min,
		       created_at, updated_at,
		       partner_id, managed_by, client_id
		FROM tenants WHERE partner_id = $1 ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, fmt.Errorf("list tenants by partner: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		t, err := scanTenantRow(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// GetStats returns usage statistics for a tenant over the last 30 days.
func (s *Service) GetStats(ctx context.Context, tenantID string, since time.Time) (conversations, leads, appointments int, err error) {
	err = s.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE event_type = 'session_started'),
			COUNT(*) FILTER (WHERE event_type = 'lead_captured'),
			COUNT(*) FILTER (WHERE event_type = 'appointment_booked')
		FROM analytics_events
		WHERE tenant_id = $1 AND created_at >= $2
	`, tenantID, since).Scan(&conversations, &leads, &appointments)
	return
}

// Create inserts a new tenant and returns the created record.
func (s *Service) Create(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	apiKey := generateAPIKey()

	bizInfoJSON, err := json.Marshal(req.BusinessInfo)
	if err != nil {
		return nil, fmt.Errorf("marshal business info: %w", err)
	}

	kbJSON, err := json.Marshal(req.KnowledgeBase)
	if err != nil {
		return nil, fmt.Errorf("marshal knowledge base: %w", err)
	}

	// Encrypt scheduler config API key if present
	schedulerCfg := req.SchedulerConfig
	if schedulerCfg.APIKey != "" {
		encKey, err := encrypt.Encrypt(schedulerCfg.APIKey, s.encryptionKey)
		if err != nil {
			s.log.Warn("failed to encrypt scheduler API key", zap.Error(err))
		} else {
			schedulerCfg.APIKey = encKey
		}
	}

	schedulerCfgJSON, err := json.Marshal(schedulerCfg)
	if err != nil {
		return nil, fmt.Errorf("marshal scheduler config: %w", err)
	}

	leadFormJSON, err := json.Marshal(req.LeadFormConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal lead form config: %w", err)
	}

	botName := req.BotName
	if botName == "" {
		botName = "Assistant"
	}
	primaryColor := req.PrimaryColor
	if primaryColor == "" {
		primaryColor = "#6C63FF"
	}
	greeting := req.Greeting
	if greeting == "" {
		greeting = "Hi! How can I help you today?"
	}
	position := req.Position
	if position == "" {
		position = "bottom-right"
	}
	plan := req.Plan
	if plan == "" {
		plan = "starter"
	}

	managedBy := req.ManagedBy
	if managedBy == "" {
		managedBy = "bpa"
	}

	var id string
	err = s.db.QueryRow(ctx, `
		INSERT INTO tenants (
			api_key, name, plan, is_active,
			bot_name, avatar_url, primary_color, greeting, position,
			business_info, knowledge_base,
			scheduler_type, scheduler_config,
			lead_capture_enabled, lead_form_config, lead_webhook_url, lead_notify_email,
			discord_webhook_url,
			partner_id, managed_by, client_id,
			portals_instance_id, product_instance_id
		) VALUES ($1,$2,$3,true,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id`,
		apiKey, req.Name, plan,
		botName, req.SchedulerConfig.APIKey, primaryColor, greeting, position,
		bizInfoJSON, kbJSON,
		req.SchedulerType, schedulerCfgJSON,
		req.LeadCaptureEnabled, leadFormJSON, req.LeadWebhookURL, req.LeadNotifyEmail,
		req.DiscordWebhookURL,
		req.PartnerID, managedBy, req.ClientID,
		req.PortalsInstanceID, nullableString(req.ProductInstanceID),
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert tenant: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Update modifies an existing tenant's configuration.
func (s *Service) Update(ctx context.Context, id string, req UpdateTenantRequest) (*Tenant, error) {
	bizInfoJSON, _ := json.Marshal(req.BusinessInfo)
	kbJSON, _ := json.Marshal(req.KnowledgeBase)

	schedulerCfg := req.SchedulerConfig
	if schedulerCfg.APIKey != "" {
		if encKey, err := encrypt.Encrypt(schedulerCfg.APIKey, s.encryptionKey); err == nil {
			schedulerCfg.APIKey = encKey
		}
	}
	schedulerCfgJSON, _ := json.Marshal(schedulerCfg)
	leadFormJSON, _ := json.Marshal(req.LeadFormConfig)

	_, err := s.db.Exec(ctx, `
		UPDATE tenants SET
			name=$2, plan=$3,
			bot_name=$4, primary_color=$5, greeting=$6, position=$7,
			business_info=$8, knowledge_base=$9,
			scheduler_type=$10, scheduler_config=$11,
			lead_capture_enabled=$12, lead_form_config=$13,
			lead_webhook_url=$14, lead_notify_email=$15,
			discord_webhook_url=$16, updated_at=$17
		WHERE id=$1`,
		id, req.Name, defaultStr(req.Plan, "starter"),
		defaultStr(req.BotName, "Assistant"), defaultStr(req.PrimaryColor, "#6C63FF"),
		defaultStr(req.Greeting, "Hi! How can I help you today?"), defaultStr(req.Position, "bottom-right"),
		bizInfoJSON, kbJSON,
		req.SchedulerType, schedulerCfgJSON,
		req.LeadCaptureEnabled, leadFormJSON,
		req.LeadWebhookURL, req.LeadNotifyEmail,
		req.DiscordWebhookURL, time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("update tenant: %w", err)
	}

	return s.GetByID(ctx, id)
}

// GetPortalsInstanceID returns the portals_instance_id for a tenant.
func (s *Service) GetPortalsInstanceID(ctx context.Context, id string) (string, error) {
	var v string
	err := s.db.QueryRow(ctx, `SELECT COALESCE(portals_instance_id,'') FROM tenants WHERE id=$1`, id).Scan(&v)
	return v, err
}

// Delete permanently removes a tenant.
func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	return err
}

// SetActive sets the is_active flag on a tenant.
func (s *Service) SetActive(ctx context.Context, id string, active bool) error {
	_, err := s.db.Exec(ctx, "UPDATE tenants SET is_active = $2, updated_at = NOW() WHERE id = $1", id, active)
	return err
}

// RotateAPIKey generates a new API key for a tenant.
func (s *Service) RotateAPIKey(ctx context.Context, id string) (string, error) {
	newKey := generateAPIKey()
	_, err := s.db.Exec(ctx, `UPDATE tenants SET api_key=$2, updated_at=$3 WHERE id=$1`,
		id, newKey, time.Now().UTC())
	if err != nil {
		return "", fmt.Errorf("rotate api key: %w", err)
	}
	return newKey, nil
}

// ToWidgetConfig converts a Tenant to its public WidgetConfig.
func (t *Tenant) ToWidgetConfig() WidgetConfig {
	placeholder := "Type a message..."
	return WidgetConfig{
		TenantID:     t.ID,
		BotName:      t.BotName,
		AvatarURL:    t.AvatarURL,
		PrimaryColor: t.PrimaryColor,
		AccentColor:  "#FFFFFF",
		Position:     t.Position,
		Greeting:     t.Greeting,
		Placeholder:  placeholder,
		Features: WidgetFeatures{
			LeadCapture:   t.LeadCaptureEnabled,
			Booking:       t.SchedulerType != "",
			SchedulerType: t.SchedulerType,
		},
		LeadForm: t.LeadFormConfig,
	}
}

// PopulateJSONFields deserializes raw JSONB fields into typed structs.
func (t *Tenant) PopulateJSONFields(encryptionKey string) error {
	if len(t.BusinessInfoRaw) > 0 {
		bizInfo, err := knowledge.ParseBusinessInfo(t.BusinessInfoRaw)
		if err != nil {
			return err
		}
		t.BusinessInfo = bizInfo
	}

	if len(t.KnowledgeBaseRaw) > 0 {
		kb, err := knowledge.ParseKnowledgeBase(t.KnowledgeBaseRaw)
		if err != nil {
			return err
		}
		t.KnowledgeBase = kb
	}

	if len(t.SchedulerConfigRaw) > 0 {
		var cfg SchedulerConfig
		if err := json.Unmarshal(t.SchedulerConfigRaw, &cfg); err == nil {
			// Decrypt API key
			if cfg.APIKey != "" && encryptionKey != "" {
				if decrypted, err := encrypt.Decrypt(cfg.APIKey, encryptionKey); err == nil {
					cfg.APIKey = decrypted
				}
			}
			t.SchedulerConfig = cfg
		}
	}

	if len(t.LeadFormConfigRaw) > 0 {
		var lfc LeadFormConfig
		if err := json.Unmarshal(t.LeadFormConfigRaw, &lfc); err == nil {
			t.LeadFormConfig = lfc
		}
	}

	return nil
}

func generateAPIKey() string {
	return "bpkey_" + uuid.New().String()
}

func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// nullableString returns nil if s is empty, otherwise a pointer to s.
// Used for optional UUID foreign key columns.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// rowScanner is implemented by both pgx.Row and pgx.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanTenant(row rowScanner) (*Tenant, error) {
	return scanTenantRow(row)
}

func scanTenantRow(row rowScanner) (*Tenant, error) {
	var t Tenant
	err := row.Scan(
		&t.ID, &t.APIKey, &t.Name, &t.Plan, &t.IsActive,
		&t.BotName, &t.AvatarURL, &t.PrimaryColor, &t.Greeting, &t.Position,
		&t.BusinessInfoRaw, &t.KnowledgeBaseRaw,
		&t.SchedulerType, &t.SchedulerConfigRaw,
		&t.LeadCaptureEnabled, &t.LeadFormConfigRaw, &t.LeadWebhookURL, &t.LeadNotifyEmail,
		&t.DiscordWebhookURL, &t.MaxMessagesPerDay, &t.MaxMessagesPerMin,
		&t.CreatedAt, &t.UpdatedAt,
		&t.PartnerID, &t.ManagedBy, &t.ClientID,
	)
	if err != nil {
		return nil, fmt.Errorf("scan tenant: %w", err)
	}
	return &t, nil
}
