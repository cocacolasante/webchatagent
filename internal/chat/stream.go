package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter wraps an http.ResponseWriter to write Server-Sent Events.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates a new SSEWriter. Returns an error if streaming is not supported.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	return &SSEWriter{w: w, flusher: flusher}, nil
}

// WriteChunk sends a text chunk event.
func (s *SSEWriter) WriteChunk(text string) error {
	data, _ := json.Marshal(map[string]string{"text": text})
	return s.writeEvent("chunk", string(data))
}

// WriteDone sends the completion event.
func (s *SSEWriter) WriteDone(sessionID string, leadPromptSuggested, bookingPromptSuggested bool) error {
	data, _ := json.Marshal(map[string]interface{}{
		"sessionId":              sessionID,
		"leadPromptSuggested":    leadPromptSuggested,
		"bookingPromptSuggested": bookingPromptSuggested,
	})
	return s.writeEvent("done", string(data))
}

// WriteError sends an error event.
func (s *SSEWriter) WriteError(message string) error {
	data, _ := json.Marshal(map[string]string{"message": message})
	return s.writeEvent("error", string(data))
}

// WriteSystem sends a system event (lead form prompt, booking prompt, etc.).
func (s *SSEWriter) WriteSystem(eventType, message string) error {
	data, _ := json.Marshal(map[string]string{
		"type":    eventType,
		"message": message,
	})
	return s.writeEvent("system", string(data))
}

func (s *SSEWriter) writeEvent(eventName, data string) error {
	_, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventName, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}
