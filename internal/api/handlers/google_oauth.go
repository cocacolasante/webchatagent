package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/blueprintautomation/blueprint-chat/internal/config"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	googleOAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL  = "https://oauth2.googleapis.com/token"
	googleCalScope  = "https://www.googleapis.com/auth/calendar"
	oauthStateTTL   = 10 * time.Minute
	oauthStatePrefix = "oauth:google:state:"
)

// GoogleOAuthHandler manages the Google Calendar OAuth 2.0 flow.
type GoogleOAuthHandler struct {
	tenantSvc *tenant.Service
	redis     *redis.Client
	cfg       *config.Config
	log       *zap.Logger
}

func NewGoogleOAuthHandler(svc *tenant.Service, rdb *redis.Client, cfg *config.Config, log *zap.Logger) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{tenantSvc: svc, redis: rdb, cfg: cfg, log: log}
}

// StartOAuth initiates the Google OAuth flow for a tenant.
// GET /api/admin/oauth/google/start?tenant_id=X
// Requires admin auth (called from the admin dashboard).
func (h *GoogleOAuthHandler) StartOAuth(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, `{"error":"tenant_id required"}`, http.StatusBadRequest)
		return
	}

	if h.cfg.GoogleClientID == "" {
		http.Error(w, `{"error":"Google OAuth not configured"}`, http.StatusServiceUnavailable)
		return
	}

	// Generate random opaque state token
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	state := hex.EncodeToString(b)

	// Store tenant_id in Redis keyed by state (10-min TTL)
	if err := h.redis.Set(r.Context(), oauthStatePrefix+state, tenantID, oauthStateTTL).Err(); err != nil {
		h.log.Error("store oauth state", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	params := url.Values{
		"client_id":     {h.cfg.GoogleClientID},
		"redirect_uri":  {h.redirectURI()},
		"response_type": {"code"},
		"scope":         {googleCalScope},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
		"state":         {state},
	}

	http.Redirect(w, r, googleOAuthURL+"?"+params.Encode(), http.StatusFound)
}

// Callback handles the redirect from Google after user consent.
// GET /api/admin/oauth/google/callback?code=X&state=X
// This endpoint is public — Google redirects here; auth is via the state token.
func (h *GoogleOAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	adminURL := h.cfg.BaseURL + "/admin/"

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		http.Redirect(w, r, adminURL+"?gcal_error="+url.QueryEscape(errParam), http.StatusFound)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		http.Redirect(w, r, adminURL+"?gcal_error=invalid_response", http.StatusFound)
		return
	}

	// Retrieve and delete tenant ID from Redis state
	tenantID, err := h.redis.GetDel(r.Context(), oauthStatePrefix+state).Result()
	if err != nil {
		http.Redirect(w, r, adminURL+"?gcal_error=state_expired", http.StatusFound)
		return
	}

	// Exchange authorization code for tokens
	tokens, err := h.exchangeCode(r.Context(), code)
	if err != nil {
		h.log.Error("exchange google oauth code", zap.Error(err))
		http.Redirect(w, r, adminURL+"?gcal_error=token_exchange_failed", http.StatusFound)
		return
	}

	// Load tenant and update scheduler config
	t, err := h.tenantSvc.GetByID(r.Context(), tenantID)
	if err != nil {
		http.Redirect(w, r, adminURL+"?gcal_error=tenant_not_found", http.StatusFound)
		return
	}

	schedulerCfg := t.SchedulerConfig
	schedulerCfg.APIKey = tokens.AccessToken
	schedulerCfg.RefreshToken = tokens.RefreshToken
	if schedulerCfg.CalendarID == "" {
		schedulerCfg.CalendarID = "primary"
	}

	updateReq := tenant.UpdateTenantRequest{
		Name:                t.Name,
		BotName:             t.BotName,
		PrimaryColor:        t.PrimaryColor,
		Greeting:            t.Greeting,
		Position:            t.Position,
		Plan:                t.Plan,
		SchedulerType:       "google",
		SchedulerConfig:     schedulerCfg,
		LeadCaptureEnabled:  t.LeadCaptureEnabled,
		LeadFormConfig:      t.LeadFormConfig,
		LeadWebhookURL:      t.LeadWebhookURL,
		LeadNotifyEmail:     t.LeadNotifyEmail,
		DiscordWebhookURL:   t.DiscordWebhookURL,
	}

	if _, err = h.tenantSvc.Update(r.Context(), tenantID, updateReq); err != nil {
		h.log.Error("update tenant google scheduler", zap.Error(err))
		http.Redirect(w, r, adminURL+"?gcal_error=save_failed", http.StatusFound)
		return
	}

	// Redirect back to the tenant detail page
	http.Redirect(w, r, adminURL+"tenants/"+tenantID+"?gcal_connected=1", http.StatusFound)
}

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
}

func (h *GoogleOAuthHandler) exchangeCode(ctx context.Context, code string) (*oauthTokenResponse, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.GoogleClientID},
		"client_secret": {h.cfg.GoogleClientSecret},
		"redirect_uri":  {h.redirectURI()},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var tokens oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if tokens.Error != "" {
		return nil, fmt.Errorf("google oauth error: %s", tokens.Error)
	}
	if tokens.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}
	return &tokens, nil
}

func (h *GoogleOAuthHandler) redirectURI() string {
	return h.cfg.BaseURL + "/api/admin/oauth/google/callback"
}
