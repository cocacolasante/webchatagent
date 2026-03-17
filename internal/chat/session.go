package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "blueprint-chat:session:"
const maxHistoryMessages = 40 // 20 turns = 40 messages

// Session represents an active chat session.
type Session struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	Messages       []Message `json:"messages"`
	LeadCaptured   bool      `json:"leadCaptured"`
	LeadID         string    `json:"leadId,omitempty"`
	BookingOffered bool      `json:"bookingOffered"`
	AppointmentID  string    `json:"appointmentId,omitempty"`
	SourceURL      string    `json:"sourceUrl"`
	UTMParams      map[string]string `json:"utmParams,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	LastActivityAt time.Time `json:"lastActivityAt"`
}

// Message is a single chat turn.
type Message struct {
	Role      string    `json:"role"` // "user" | "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// SessionManager handles Redis-backed session storage.
type SessionManager struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(redisClient *redis.Client, ttl time.Duration) *SessionManager {
	return &SessionManager{redis: redisClient, ttl: ttl}
}

// GetOrCreate retrieves an existing session or creates a new one.
func (m *SessionManager) GetOrCreate(ctx context.Context, sessionID, tenantID, sourceURL string, utmParams map[string]string) (*Session, bool, error) {
	if sessionID != "" {
		if sess, err := m.Get(ctx, sessionID); err == nil {
			return sess, false, nil
		}
	}

	// Create new session
	sess := &Session{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		Messages:       []Message{},
		SourceURL:      sourceURL,
		UTMParams:      utmParams,
		CreatedAt:      time.Now().UTC(),
		LastActivityAt: time.Now().UTC(),
	}

	if err := m.Save(ctx, sess); err != nil {
		return nil, false, fmt.Errorf("save new session: %w", err)
	}

	return sess, true, nil
}

// Get retrieves a session by ID.
func (m *SessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
	data, err := m.redis.Get(ctx, sessionKeyPrefix+sessionID).Bytes()
	if err != nil {
		return nil, fmt.Errorf("get session %s: %w", sessionID, err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &sess, nil
}

// Save persists a session to Redis and refreshes its TTL.
func (m *SessionManager) Save(ctx context.Context, sess *Session) error {
	sess.LastActivityAt = time.Now().UTC()

	// Trim to max history
	if len(sess.Messages) > maxHistoryMessages {
		sess.Messages = sess.Messages[len(sess.Messages)-maxHistoryMessages:]
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	return m.redis.Set(ctx, sessionKeyPrefix+sess.ID, data, m.ttl).Err()
}

// AddMessage appends a message to the session and saves.
func (m *SessionManager) AddMessage(ctx context.Context, sess *Session, role, content string) error {
	sess.Messages = append(sess.Messages, Message{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	})
	return m.Save(ctx, sess)
}

// Delete removes a session from Redis.
func (m *SessionManager) Delete(ctx context.Context, sessionID string) error {
	return m.redis.Del(ctx, sessionKeyPrefix+sessionID).Err()
}
