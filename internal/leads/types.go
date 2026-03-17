package leads

import "time"

// Lead represents a captured contact.
type Lead struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	ConversationID string    `json:"conversationId,omitempty"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName,omitempty"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone,omitempty"`
	SourceURL      string    `json:"sourceUrl,omitempty"`
	UTMSource      string    `json:"utmSource,omitempty"`
	UTMMedium      string    `json:"utmMedium,omitempty"`
	UTMCampaign    string    `json:"utmCampaign,omitempty"`
	SessionSummary string    `json:"sessionSummary,omitempty"`
	Status         string    `json:"status"` // new | contacted | qualified | closed
	WebhookSent    bool      `json:"webhookSent"`
	WebhookSentAt  *time.Time `json:"webhookSentAt,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// CaptureRequest is the lead form submission payload.
type CaptureRequest struct {
	FirstName      string `json:"firstName" validate:"required"`
	LastName       string `json:"lastName"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone"`
	SourceURL      string `json:"sourceUrl"`
	UTMSource      string `json:"utmSource"`
	UTMMedium      string `json:"utmMedium"`
	UTMCampaign    string `json:"utmCampaign"`
	SessionID      string `json:"sessionId"`
	SessionSummary string `json:"sessionSummary"`
}

// WebhookPayload is the outgoing webhook body on lead capture.
type WebhookPayload struct {
	Event    string `json:"event"`
	TenantID string `json:"tenantId"`
	Lead     Lead   `json:"lead"`
}
