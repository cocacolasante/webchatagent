package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DiscordSender sends messages to Discord webhooks.
type DiscordSender struct {
	client *http.Client
	log    *zap.Logger
}

// NewDiscordSender creates a new DiscordSender.
func NewDiscordSender(log *zap.Logger) *DiscordSender {
	return &DiscordSender{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    log,
	}
}

// Send posts a message to a Discord webhook URL.
func (d *DiscordSender) Send(ctx context.Context, webhookURL, content string) error {
	if webhookURL == "" {
		return nil
	}

	payload := map[string]string{"content": content}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("discord request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook error: status %d", resp.StatusCode)
	}

	return nil
}
