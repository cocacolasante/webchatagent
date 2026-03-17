package plugins

import (
	"context"

	"github.com/blueprintautomation/blueprint-chat/internal/chat"
	"github.com/blueprintautomation/blueprint-chat/internal/leads"
	"github.com/blueprintautomation/blueprint-chat/internal/scheduler"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// Registry holds all registered plugins.
type Registry struct {
	plugins []Plugin
	log     *zap.Logger
}

// NewRegistry creates a new plugin Registry.
func NewRegistry(log *zap.Logger) *Registry {
	return &Registry{log: log}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(p Plugin) {
	r.plugins = append(r.plugins, p)
	r.log.Info("plugin registered", zap.String("name", p.Name()), zap.String("version", p.Version()))
}

// EnabledForTenant returns all registered plugins.
// In the future, this could filter by tenant-specific plugin config.
func (r *Registry) EnabledForTenant(tenantID string) []Plugin {
	return r.plugins
}

// FireOnMessageReceived triggers all plugins' OnMessageReceived hooks.
func (r *Registry) FireOnMessageReceived(ctx context.Context, msg *chat.Message, t *tenant.Tenant) {
	for _, p := range r.EnabledForTenant(t.ID) {
		if err := p.OnMessageReceived(ctx, msg, t); err != nil {
			r.log.Error("plugin OnMessageReceived error",
				zap.String("plugin", p.Name()), zap.Error(err))
		}
	}
}

// FireOnResponseGenerated triggers all plugins' OnResponseGenerated hooks.
func (r *Registry) FireOnResponseGenerated(ctx context.Context, resp *ChatResponse, t *tenant.Tenant) {
	for _, p := range r.EnabledForTenant(t.ID) {
		if err := p.OnResponseGenerated(ctx, resp, t); err != nil {
			r.log.Error("plugin OnResponseGenerated error",
				zap.String("plugin", p.Name()), zap.Error(err))
		}
	}
}

// FireOnLeadCaptured triggers all plugins' OnLeadCaptured hooks.
func (r *Registry) FireOnLeadCaptured(ctx context.Context, lead *leads.Lead, t *tenant.Tenant) {
	for _, p := range r.EnabledForTenant(t.ID) {
		if err := p.OnLeadCaptured(ctx, lead, t); err != nil {
			r.log.Error("plugin OnLeadCaptured error",
				zap.String("plugin", p.Name()), zap.Error(err))
		}
	}
}

// FireOnAppointmentBooked triggers all plugins' OnAppointmentBooked hooks.
func (r *Registry) FireOnAppointmentBooked(ctx context.Context, appt *scheduler.Appointment, t *tenant.Tenant) {
	for _, p := range r.EnabledForTenant(t.ID) {
		if err := p.OnAppointmentBooked(ctx, appt, t); err != nil {
			r.log.Error("plugin OnAppointmentBooked error",
				zap.String("plugin", p.Name()), zap.Error(err))
		}
	}
}

// FireOnSessionStarted triggers all plugins' OnSessionStarted hooks.
func (r *Registry) FireOnSessionStarted(ctx context.Context, session *chat.Session, t *tenant.Tenant) {
	for _, p := range r.EnabledForTenant(t.ID) {
		if err := p.OnSessionStarted(ctx, session, t); err != nil {
			r.log.Error("plugin OnSessionStarted error",
				zap.String("plugin", p.Name()), zap.Error(err))
		}
	}
}

// FireOnSessionEnded triggers all plugins' OnSessionEnded hooks.
func (r *Registry) FireOnSessionEnded(ctx context.Context, session *chat.Session, t *tenant.Tenant) {
	for _, p := range r.EnabledForTenant(t.ID) {
		if err := p.OnSessionEnded(ctx, session, t); err != nil {
			r.log.Error("plugin OnSessionEnded error",
				zap.String("plugin", p.Name()), zap.Error(err))
		}
	}
}
