package handlers

import (
	"encoding/json"
	"net/http"

	mw "github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/leads"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// LeadsHandler handles lead capture requests.
type LeadsHandler struct {
	service    *leads.Service
	tenantsSvc *tenant.Service
	log        *zap.Logger
}

// NewLeadsHandler creates a new LeadsHandler.
func NewLeadsHandler(service *leads.Service, tenantsSvc *tenant.Service, log *zap.Logger) *LeadsHandler {
	return &LeadsHandler{service: service, tenantsSvc: tenantsSvc, log: log}
}

// Capture handles POST /api/leads — captures a lead from the widget form.
func (h *LeadsHandler) Capture(w http.ResponseWriter, r *http.Request) {
	t := mw.TenantFromContext(r.Context())
	if t == nil {
		http.Error(w, `{"error":"tenant required"}`, http.StatusBadRequest)
		return
	}

	if !t.LeadCaptureEnabled {
		http.Error(w, `{"error":"lead capture is not enabled"}`, http.StatusForbidden)
		return
	}

	var req leads.CaptureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	lead, err := h.service.Capture(r.Context(), t.ID, req)
	if err != nil {
		h.log.Error("lead capture failed", zap.Error(err), zap.String("tenantId", t.ID))
		http.Error(w, `{"error":"failed to capture lead"}`, http.StatusInternalServerError)
		return
	}

	// Dispatch webhook asynchronously
	h.service.DispatchWebhook(lead, t.LeadWebhookURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"leadId":  lead.ID,
	})
}

// List handles GET /api/admin/tenants/:id/leads — lists leads for a tenant.
func (h *LeadsHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := r.PathValue("id")
	if tenantID == "" {
		http.Error(w, `{"error":"tenant ID required"}`, http.StatusBadRequest)
		return
	}

	// Partner scope check: verify partner owns this tenant
	if partnerID, scoped := mw.GetPartnerScope(r.Context()); scoped {
		t, err := h.tenantsSvc.GetByID(r.Context(), tenantID)
		if err != nil || t.PartnerID == nil || *t.PartnerID != partnerID {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
	}

	leadsList, err := h.service.List(r.Context(), tenantID, 50, 0)
	if err != nil {
		h.log.Error("list leads failed", zap.Error(err))
		http.Error(w, `{"error":"failed to list leads"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"leads": leadsList,
		"total": len(leadsList),
	})
}
