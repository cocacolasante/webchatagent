package notification

// EmailMessage is the parameters for sending an email.
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// DiscordMessage is the parameters for a Discord webhook message.
type DiscordMessage struct {
	WebhookURL string
	Content    string
}
