package handlers

import (
	"encoding/json"
	"net/http"

	mw "github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/chat"
	"go.uber.org/zap"
)

// ChatHandler handles chat streaming requests.
type ChatHandler struct {
	engine        *chat.Engine
	encryptionKey string
	log           *zap.Logger
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(engine *chat.Engine, encryptionKey string, log *zap.Logger) *ChatHandler {
	return &ChatHandler{engine: engine, encryptionKey: encryptionKey, log: log}
}

// ChatRequest is the request body for a chat message.
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId,omitempty"`
	SourceURL string `json:"sourceUrl,omitempty"`
}

// Stream handles POST /api/chat/stream — streams a Claude response via SSE.
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	t := mw.TenantFromContext(r.Context())
	if t == nil {
		http.Error(w, `{"error":"tenant required"}`, http.StatusBadRequest)
		return
	}

	// Populate parsed JSON fields
	if err := t.PopulateJSONFields(h.encryptionKey); err != nil {
		h.log.Error("populate tenant fields", zap.Error(err), zap.String("tenantId", t.ID))
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, `{"error":"message is required"}`, http.StatusBadRequest)
		return
	}

	// Parse UTM params from query string
	utmParams := map[string]string{}
	for _, key := range []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"} {
		if v := r.URL.Query().Get(key); v != "" {
			utmParams[key] = v
		}
	}

	streamReq := chat.StreamRequest{
		SessionID: req.SessionID,
		TenantID:  t.ID,
		Message:   req.Message,
		SourceURL: req.SourceURL,
		UTMParams: utmParams,
	}

	if err := h.engine.Stream(r.Context(), w, t, streamReq); err != nil {
		h.log.Error("chat stream error", zap.Error(err), zap.String("tenantId", t.ID))
	}
}
