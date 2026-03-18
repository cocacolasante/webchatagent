package tenant

import (
	"encoding/json"
	"time"

	"github.com/blueprintautomation/blueprint-chat/internal/knowledge"
)

// Tenant is the full tenant record from PostgreSQL.
type Tenant struct {
	ID      string `json:"id" db:"id"`
	APIKey  string `json:"apiKey" db:"api_key"`
	Name    string `json:"name" db:"name"`
	Plan    string `json:"plan" db:"plan"`
	IsActive bool  `json:"isActive" db:"is_active"`

	// Branding
	BotName      string `json:"botName" db:"bot_name"`
	AvatarURL    string `json:"avatarUrl" db:"avatar_url"`
	PrimaryColor string `json:"primaryColor" db:"primary_color"`
	Greeting     string `json:"greeting" db:"greeting"`
	Position     string `json:"position" db:"position"`

	// Business context (raw JSONB)
	BusinessInfoRaw  json.RawMessage `json:"-" db:"business_info"`
	KnowledgeBaseRaw json.RawMessage `json:"-" db:"knowledge_base"`

	// Parsed fields (populated after DB load)
	BusinessInfo  knowledge.BusinessInfo  `json:"businessInfo"`
	KnowledgeBase knowledge.KnowledgeBase `json:"knowledgeBase"`

	// Scheduler
	SchedulerType      string          `json:"schedulerType" db:"scheduler_type"`
	SchedulerConfigRaw json.RawMessage `json:"-" db:"scheduler_config"`
	SchedulerConfig    SchedulerConfig `json:"schedulerConfig"`

	// Lead capture
	LeadCaptureEnabled bool            `json:"leadCaptureEnabled" db:"lead_capture_enabled"`
	LeadFormConfigRaw  json.RawMessage `json:"-" db:"lead_form_config"`
	LeadFormConfig     LeadFormConfig  `json:"leadFormConfig"`
	LeadWebhookURL     string          `json:"leadWebhookUrl" db:"lead_webhook_url"`
	LeadNotifyEmail    string          `json:"leadNotifyEmail" db:"lead_notify_email"`

	// Notifications
	DiscordWebhookURL string `json:"discordWebhookUrl" db:"discord_webhook_url"`

	// Rate limits
	MaxMessagesPerDay int `json:"maxMessagesPerDay" db:"max_messages_per_day"`
	MaxMessagesPerMin int `json:"maxMessagesPerMin" db:"max_messages_per_min"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

	// Blueprint Command reseller support
	PartnerID *string `json:"partner_id,omitempty" db:"partner_id"`
	ManagedBy string  `json:"managed_by" db:"managed_by"`
	ClientID  *string `json:"client_id,omitempty" db:"client_id"`

	// Blueprint Command billing
	PortalsInstanceID string `json:"portalsInstanceId,omitempty" db:"portals_instance_id"`
}

// SchedulerConfig holds provider-specific scheduling configuration.
type SchedulerConfig struct {
	APIKey      string `json:"apiKey"`
	EventTypeID string `json:"eventTypeId"`
	Username    string `json:"username,omitempty"` // Calendly
	CalendarID  string `json:"calendarId,omitempty"` // Google
}

// LeadFormConfig defines the lead capture form settings.
type LeadFormConfig struct {
	Fields         []string `json:"fields"`
	RequiredFields []string `json:"requiredFields"`
	TriggerMessage string   `json:"triggerMessage"`
}

// WidgetConfig is the public-facing config returned to the widget.
type WidgetConfig struct {
	TenantID    string `json:"tenantId"`
	BotName     string `json:"botName"`
	AvatarURL   string `json:"avatarUrl"`
	PrimaryColor string `json:"primaryColor"`
	AccentColor  string `json:"accentColor"`
	Position    string `json:"position"`
	Greeting    string `json:"greeting"`
	Placeholder string `json:"placeholderText"`
	Features    WidgetFeatures `json:"features"`
	LeadForm    LeadFormConfig `json:"leadForm"`
}

// WidgetFeatures describes which features are enabled for the widget.
type WidgetFeatures struct {
	LeadCapture   bool   `json:"leadCapture"`
	Booking       bool   `json:"booking"`
	SchedulerType string `json:"schedulerType,omitempty"`
}

// CreateTenantRequest is the payload for creating a new tenant.
type CreateTenantRequest struct {
	Name                string                  `json:"name" validate:"required"`
	BotName             string                  `json:"botName"`
	PrimaryColor        string                  `json:"primaryColor"`
	Greeting            string                  `json:"greeting"`
	Position            string                  `json:"position"`
	Plan                string                  `json:"plan"`
	BusinessInfo        knowledge.BusinessInfo  `json:"businessInfo"`
	KnowledgeBase       knowledge.KnowledgeBase `json:"knowledgeBase"`
	SchedulerType       string                  `json:"schedulerType"`
	SchedulerConfig     SchedulerConfig         `json:"schedulerConfig"`
	LeadCaptureEnabled  bool                    `json:"leadCaptureEnabled"`
	LeadFormConfig      LeadFormConfig          `json:"leadFormConfig"`
	LeadWebhookURL      string                  `json:"leadWebhookUrl"`
	LeadNotifyEmail     string                  `json:"leadNotifyEmail"`
	DiscordWebhookURL   string                  `json:"discordWebhookUrl"`

	// Blueprint Command reseller fields
	PartnerID *string `json:"partner_id,omitempty"`
	ManagedBy string  `json:"managed_by,omitempty"`
	ClientID  *string `json:"client_id,omitempty"`
	PortalsInstanceID string `json:"portalsInstanceId,omitempty"`
}

// UpdateTenantRequest is the payload for updating a tenant.
type UpdateTenantRequest = CreateTenantRequest
