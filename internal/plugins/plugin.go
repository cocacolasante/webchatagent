package plugins

import (
	"context"

	"github.com/blueprintautomation/blueprint-chat/internal/chat"
	"github.com/blueprintautomation/blueprint-chat/internal/leads"
	"github.com/blueprintautomation/blueprint-chat/internal/scheduler"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
)

// ChatResponse is the full response from a chat turn.
type ChatResponse struct {
	SessionID    string
	FullText     string
	LeadPrompt   bool
	BookingPrompt bool
}

// Plugin is the interface all Blueprint Chat plugins must implement.
// Return nil from any hook to skip processing.
type Plugin interface {
	Name() string
	Version() string

	OnMessageReceived(ctx context.Context, msg *chat.Message, t *tenant.Tenant) error
	OnResponseGenerated(ctx context.Context, resp *ChatResponse, t *tenant.Tenant) error
	OnLeadCaptured(ctx context.Context, lead *leads.Lead, t *tenant.Tenant) error
	OnAppointmentBooked(ctx context.Context, appt *scheduler.Appointment, t *tenant.Tenant) error
	OnSessionStarted(ctx context.Context, session *chat.Session, t *tenant.Tenant) error
	OnSessionEnded(ctx context.Context, session *chat.Session, t *tenant.Tenant) error
}
