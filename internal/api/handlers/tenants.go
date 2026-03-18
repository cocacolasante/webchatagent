package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/billing"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// TenantsHandler handles tenant admin CRUD.
type TenantsHandler struct {
	service     *tenant.Service
	provisioner *tenant.Provisioner
	billing     *billing.TenantLimitClient
	log         *zap.Logger
}

// NewTenantsHandler creates a new TenantsHandler.
func NewTenantsHandler(svc *tenant.Service, prov *tenant.Provisioner, log *zap.Logger, billingClient *billing.TenantLimitClient) *TenantsHandler {
	return &TenantsHandler{service: svc, provisioner: prov, billing: billingClient, log: log}
}

// Create handles POST /api/admin/tenants.
func (h *TenantsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req tenant.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if !h.billing.CheckTenantAllowed(r.Context(), req.PortalsInstanceID) {
		http.Error(w, `{"error":"tenant limit reached"}`, 402)
		return
	}

	result, err := h.provisioner.Provision(r.Context(), req)
	if err != nil {
		h.log.Error("provision tenant", zap.Error(err))
		http.Error(w, `{"error":"failed to create tenant"}`, http.StatusInternalServerError)
		return
	}

	go h.billing.IncrementTenantCount(context.Background(), req.PortalsInstanceID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(result)
}

// List handles GET /api/admin/tenants.
func (h *TenantsHandler) List(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.service.List(r.Context())
	if err != nil {
		h.log.Error("list tenants", zap.Error(err))
		http.Error(w, `{"error":"failed to list tenants"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tenants": tenants,
		"total":   len(tenants),
	})
}

// Get handles GET /api/admin/tenants/:id.
func (h *TenantsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

// Update handles PUT /api/admin/tenants/:id.
func (h *TenantsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req tenant.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	t, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		h.log.Error("update tenant", zap.Error(err))
		http.Error(w, `{"error":"failed to update tenant"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

// Delete handles DELETE /api/admin/tenants/:id.
func (h *TenantsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	instanceID, _ := h.service.GetPortalsInstanceID(r.Context(), id)

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.log.Error("delete tenant", zap.Error(err))
		http.Error(w, `{"error":"failed to delete tenant"}`, http.StatusInternalServerError)
		return
	}

	go h.billing.DecrementTenantCount(context.Background(), instanceID)

	w.WriteHeader(http.StatusNoContent)
}

// RotateKey handles POST /api/admin/tenants/:id/rotate-key.
func (h *TenantsHandler) RotateKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	newKey, err := h.service.RotateAPIKey(r.Context(), id)
	if err != nil {
		h.log.Error("rotate API key", zap.Error(err))
		http.Error(w, `{"error":"failed to rotate API key"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"apiKey": newKey})
}

// GetStats handles GET /api/admin/tenants/{id}/stats
// Called by blueprint-command to populate client portal usage summaries.
func (h *TenantsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	// Partner scope check: verify partner owns this tenant
	partnerID, scoped := middleware.GetPartnerScope(ctx)
	if scoped {
		t, err := h.service.GetByID(ctx, id)
		if err != nil || t.PartnerID == nil || *t.PartnerID != partnerID {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
	}

	now := time.Now().UTC()
	periodStart := now.AddDate(0, -1, 0)

	conversationsTotal, leadsCaptured, appointmentsBooked, _ := h.service.GetStats(ctx, id, periodStart)

	summary := fmt.Sprintf(
		"Your AI chat agent handled %d conversations, captured %d leads, and booked %d appointments this month.",
		conversationsTotal, leadsCaptured, appointmentsBooked,
	)

	resp := map[string]interface{}{
		"product_key":  "webchatagent",
		"tenant_id":    id,
		"period":       "last_30_days",
		"period_start": periodStart,
		"period_end":   now,
		"stats": map[string]interface{}{
			"conversations_total": conversationsTotal,
			"leads_captured":      leadsCaptured,
			"appointments_booked": appointmentsBooked,
		},
		"summary": summary,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// UpdateKnowledge handles POST /api/admin/tenants/:id/knowledge.
func (h *TenantsHandler) UpdateKnowledge(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Get current tenant
	current, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
		return
	}

	// Decode partial knowledge update
	var req tenant.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Preserve existing values, only update what's sent
	if req.Name == "" {
		req.Name = current.Name
	}

	t, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		h.log.Error("update knowledge", zap.Error(err))
		http.Error(w, `{"error":"failed to update knowledge base"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}
