// Package example provides a reference Plugin implementation.
// Copy this file as a starting point for new plugins.
package example

import (
	"context"

	"github.com/blueprintautomation/blueprint-chat/internal/chat"
	"github.com/blueprintautomation/blueprint-chat/internal/leads"
	"github.com/blueprintautomation/blueprint-chat/internal/plugins"
	"github.com/blueprintautomation/blueprint-chat/internal/scheduler"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// ExamplePlugin is a minimal plugin that logs all lifecycle events.
// It implements the plugins.Plugin interface.
type ExamplePlugin struct {
	log *zap.Logger
}

// NewExamplePlugin creates a new ExamplePlugin.
func NewExamplePlugin(log *zap.Logger) *ExamplePlugin {
	return &ExamplePlugin{log: log}
}

// Compile-time interface check.
var _ plugins.Plugin = (*ExamplePlugin)(nil)

func (p *ExamplePlugin) Name() string    { return "example" }
func (p *ExamplePlugin) Version() string { return "1.0.0" }

func (p *ExamplePlugin) OnMessageReceived(ctx context.Context, msg *chat.Message, t *tenant.Tenant) error {
	p.log.Debug("example: message received",
		zap.String("tenant", t.ID),
		zap.String("role", msg.Role),
		zap.Int("contentLen", len(msg.Content)),
	)
	return nil
}

func (p *ExamplePlugin) OnResponseGenerated(ctx context.Context, resp *plugins.ChatResponse, t *tenant.Tenant) error {
	p.log.Debug("example: response generated",
		zap.String("tenant", t.ID),
		zap.Int("responseLen", len(resp.FullText)),
	)
	return nil
}

func (p *ExamplePlugin) OnLeadCaptured(ctx context.Context, lead *leads.Lead, t *tenant.Tenant) error {
	p.log.Info("example: lead captured",
		zap.String("tenant", t.ID),
		zap.String("email", lead.Email),
	)
	return nil
}

func (p *ExamplePlugin) OnAppointmentBooked(ctx context.Context, appt *scheduler.Appointment, t *tenant.Tenant) error {
	p.log.Info("example: appointment booked",
		zap.String("tenant", t.ID),
		zap.String("appointmentId", appt.ID),
	)
	return nil
}

func (p *ExamplePlugin) OnSessionStarted(ctx context.Context, session *chat.Session, t *tenant.Tenant) error {
	p.log.Debug("example: session started",
		zap.String("tenant", t.ID),
		zap.String("sessionId", session.ID),
	)
	return nil
}

func (p *ExamplePlugin) OnSessionEnded(ctx context.Context, session *chat.Session, t *tenant.Tenant) error {
	p.log.Debug("example: session ended",
		zap.String("tenant", t.ID),
		zap.String("sessionId", session.ID),
		zap.Int("messageCount", len(session.Messages)),
	)
	return nil
}
