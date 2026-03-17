package leads

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// LeadNotifier implements Notifier using email and Discord.
type LeadNotifier struct {
	log *zap.Logger
	// Email and Discord notifiers are injected via the notification package
	emailNotify   EmailSender
	discordNotify DiscordSender
}

// EmailSender can send emails.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// DiscordSender can send Discord webhooks.
type DiscordSender interface {
	Send(ctx context.Context, webhookURL, message string) error
}

// NewLeadNotifier creates a new LeadNotifier.
func NewLeadNotifier(log *zap.Logger, email EmailSender, discord DiscordSender) *LeadNotifier {
	return &LeadNotifier{log: log, emailNotify: email, discordNotify: discord}
}

// NotifyLead sends lead notifications asynchronously.
func (n *LeadNotifier) NotifyLead(ctx context.Context, lead *Lead, tenantName, webhookURL, email, discordURL string) error {
	// Send email notification
	if email != "" && n.emailNotify != nil {
		go func() {
			subject := fmt.Sprintf("New Lead: %s %s — %s", lead.FirstName, lead.LastName, tenantName)
			body := fmt.Sprintf(
				"New lead captured for %s\n\nName: %s %s\nEmail: %s\nPhone: %s\nSource: %s\n\nSummary: %s",
				tenantName, lead.FirstName, lead.LastName, lead.Email, lead.Phone,
				lead.SourceURL, lead.SessionSummary,
			)
			if err := n.emailNotify.Send(context.Background(), email, subject, body); err != nil {
				n.log.Error("send lead email notification", zap.Error(err))
			}
		}()
	}

	// Send Discord notification
	if discordURL != "" && n.discordNotify != nil {
		go func() {
			message := fmt.Sprintf(
				"**New Lead for %s**\n👤 %s %s\n📧 %s\n📞 %s\n🔗 %s\n\n_%s_",
				tenantName, lead.FirstName, lead.LastName, lead.Email, lead.Phone,
				lead.SourceURL, lead.SessionSummary,
			)
			if err := n.discordNotify.Send(context.Background(), discordURL, message); err != nil {
				n.log.Error("send lead discord notification", zap.Error(err))
			}
		}()
	}

	return nil
}
