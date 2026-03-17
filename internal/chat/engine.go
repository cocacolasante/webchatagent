package chat

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/blueprintautomation/blueprint-chat/internal/knowledge"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// Engine is the core chat engine that streams Claude responses.
type Engine struct {
	anthropic  *anthropic.Client
	sessions   *SessionManager
	assembler  *knowledge.Assembler
	model      string
	maxTokens  int64
	log        *zap.Logger
}

// NewEngine creates a new chat Engine.
func NewEngine(
	anthropicClient *anthropic.Client,
	sessions *SessionManager,
	assembler *knowledge.Assembler,
	model string,
	maxTokens int,
	log *zap.Logger,
) *Engine {
	return &Engine{
		anthropic: anthropicClient,
		sessions:  sessions,
		assembler: assembler,
		model:     model,
		maxTokens: int64(maxTokens),
		log:       log,
	}
}

// StreamRequest is the input for a chat stream request.
type StreamRequest struct {
	SessionID  string
	TenantID   string
	Message    string
	SourceURL  string
	UTMParams  map[string]string
}

// StreamResponse is the result metadata after streaming completes.
type StreamResponse struct {
	SessionID              string
	FullResponse           string
	LeadPromptSuggested    bool
	BookingPromptSuggested bool
}

// Stream generates a Claude response and writes it to the SSE writer.
func (e *Engine) Stream(ctx context.Context, w http.ResponseWriter, t *tenant.Tenant, req StreamRequest) error {
	sse, err := NewSSEWriter(w)
	if err != nil {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return err
	}

	// Get or create session
	sess, _, err := e.sessions.GetOrCreate(ctx, req.SessionID, req.TenantID, req.SourceURL, req.UTMParams)
	if err != nil {
		_ = sse.WriteError("Session error. Please refresh and try again.")
		return fmt.Errorf("get/create session: %w", err)
	}

	// Build system prompt
	systemPrompt, err := e.assembler.BuildSystemPrompt(ctx, t.ID, t.BusinessInfo, t.KnowledgeBase)
	if err != nil {
		e.log.Error("failed to build system prompt", zap.Error(err))
		systemPrompt = "You are a helpful AI assistant."
	}

	// Build message history for Claude
	messages := e.buildMessages(sess, req.Message)

	// Save user message to session
	if err := e.sessions.AddMessage(ctx, sess, "user", req.Message); err != nil {
		e.log.Error("failed to save user message", zap.Error(err))
	}

	// Stream from Claude
	var responseBuilder strings.Builder

	stream := e.anthropic.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(e.model),
		MaxTokens: anthropic.F(e.maxTokens),
		System: anthropic.F([]anthropic.TextBlockParam{
			{Text: anthropic.F(systemPrompt)},
		}),
		Messages: anthropic.F(messages),
	})

	for stream.Next() {
		event := stream.Current()
		switch delta := event.Delta.(type) {
		case anthropic.ContentBlockDeltaEventDelta:
			if delta.Type == anthropic.ContentBlockDeltaEventDeltaTypeTextDelta {
				text := delta.Text
				responseBuilder.WriteString(text)
				if err := sse.WriteChunk(text); err != nil {
					e.log.Warn("sse write chunk error", zap.Error(err))
					return nil
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		e.log.Error("anthropic stream error", zap.Error(err))
		_ = sse.WriteError("I'm having trouble responding right now. Please try again in a moment.")
		return nil
	}

	fullResponse := responseBuilder.String()

	// Save assistant response to session
	if err := e.sessions.AddMessage(ctx, sess, "assistant", fullResponse); err != nil {
		e.log.Error("failed to save assistant message", zap.Error(err))
	}

	// Detect whether to suggest lead capture or booking
	leadPrompt, bookingPrompt := e.detectIntents(req.Message, fullResponse, sess)

	// Send done event
	_ = sse.WriteDone(sess.ID, leadPrompt, bookingPrompt)

	return nil
}

// buildMessages converts session history + current message to Anthropic format.
func (e *Engine) buildMessages(sess *Session, currentMessage string) []anthropic.MessageParam {
	var messages []anthropic.MessageParam

	for _, msg := range sess.Messages {
		if msg.Role == "assistant" {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		} else {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	// Add current message
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(currentMessage)))

	return messages
}

// detectIntents uses simple heuristics to detect lead capture / booking intent.
func (e *Engine) detectIntents(userMsg, assistantMsg string, sess *Session) (leadPrompt, bookingPrompt bool) {
	if sess.LeadCaptured {
		return false, false
	}

	lowerUser := strings.ToLower(userMsg)
	lowerAssistant := strings.ToLower(assistantMsg)
	combined := lowerUser + " " + lowerAssistant

	// Lead capture signals
	leadKeywords := []string{
		"pricing", "price", "cost", "how much", "quote", "interested",
		"contact me", "reach out", "get in touch", "follow up", "more information",
		"tell me more", "sign up", "get started",
	}
	for _, kw := range leadKeywords {
		if strings.Contains(combined, kw) {
			leadPrompt = true
			break
		}
	}

	// Booking signals
	if !sess.BookingOffered {
		bookingKeywords := []string{
			"schedule", "book", "appointment", "meeting", "call", "demo",
			"consult", "talk to someone", "speak with", "availability",
			"calendar", "time slot", "when can", "set up a",
		}
		for _, kw := range bookingKeywords {
			if strings.Contains(combined, kw) {
				bookingPrompt = true
				break
			}
		}
	}

	return leadPrompt, bookingPrompt
}
