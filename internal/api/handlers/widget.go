package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	mw "github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"go.uber.org/zap"
)

// WidgetHandler serves the widget bundle and widget config.
type WidgetHandler struct {
	widgetPath string
	log        *zap.Logger
}

// NewWidgetHandler creates a new WidgetHandler.
func NewWidgetHandler(widgetPath string, log *zap.Logger) *WidgetHandler {
	return &WidgetHandler{widgetPath: widgetPath, log: log}
}

// ServeWidget serves the compiled widget.js bundle.
func (h *WidgetHandler) ServeWidget(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(h.widgetPath)
	if err != nil {
		h.log.Error("widget bundle not found", zap.String("path", h.widgetPath), zap.Error(err))
		http.Error(w, "Widget not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(data)
}

// ServeConfig returns the widget configuration JSON for a tenant.
func (h *WidgetHandler) ServeConfig(w http.ResponseWriter, r *http.Request) {
	t := mw.TenantFromContext(r.Context())
	if t == nil {
		http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
		return
	}

	if err := t.PopulateJSONFields(""); err != nil {
		h.log.Error("populate tenant fields", zap.Error(err))
	}

	config := t.ToWidgetConfig()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=60")
	_ = json.NewEncoder(w).Encode(config)
}
