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
	instanceSvc *tenant.InstanceService
	log         *zap.Logger
}

// NewTenantsHandler creates a new TenantsHandler.
func NewTenantsHandler(svc *tenant.Service, prov *tenant.Provisioner, log *zap.Logger, billingClient *billing.TenantLimitClient, instanceSvc *tenant.InstanceService) *TenantsHandler {
	return &TenantsHandler{service: svc, provisioner: prov, billing: billingClient, instanceSvc: instanceSvc, log: log}
}

// Create handles POST /api/admin/tenants.
func (h *TenantsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req tenant.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Stamp partner_id when called with a partner key
	if partnerID, scoped := middleware.GetPartnerScope(r.Context()); scoped {
		req.PartnerID = &partnerID
	}

	if !h.billing.CheckTenantAllowed(r.Context(), req.PortalsInstanceID) {
		http.Error(w, `{"error":"tenant limit reached"}`, 402)
		return
	}

	// Determine product instance
	var inst *tenant.ProductInstance
	var err error
	if req.ProductInstanceID != "" {
		inst, err = h.instanceSvc.GetByID(r.Context(), req.ProductInstanceID)
		if err != nil {
			http.Error(w, `{"error":"product instance not found"}`, http.StatusBadRequest)
			return
		}
	} else {
		inst, err = h.instanceSvc.EnsureDefault(r.Context(), req.PortalsInstanceID, req.PortalsInstanceID)
		if err != nil {
			h.log.Error("resolve product instance", zap.Error(err))
			http.Error(w, `{"error":"failed to resolve product instance"}`, http.StatusInternalServerError)
			return
		}
	}

	// Enforce 15-tenant cap
	if inst.TenantCount >= tenant.MaxTenantsPerInstance {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         "tenant_limit_reached",
			"message":       "This instance has reached the maximum of 15 tenants. Provision a second instance to add more.",
			"current_count": inst.TenantCount,
			"max_allowed":   tenant.MaxTenantsPerInstance,
		})
		return
	}

	req.ProductInstanceID = inst.ID

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

// checkPartnerOwns returns true if the caller is not partner-scoped, or if
// the tenant with the given id belongs to the caller's partner.
func (h *TenantsHandler) checkPartnerOwns(ctx context.Context, id string) bool {
	partnerID, scoped := middleware.GetPartnerScope(ctx)
	if !scoped {
		return true
	}
	t, err := h.service.GetByID(ctx, id)
	return err == nil && t.PartnerID != nil && *t.PartnerID == partnerID
}

// List handles GET /api/admin/tenants.
func (h *TenantsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	partnerID, scoped := middleware.GetPartnerScope(ctx)

	var tenants []*tenant.Tenant
	var err error
	if scoped {
		tenants, err = h.service.ListByPartner(ctx, partnerID)
	} else {
		tenants, err = h.service.List(ctx)
	}
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
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
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
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

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
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	instanceID, _ := h.service.GetPortalsInstanceID(r.Context(), id)

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.log.Error("delete tenant", zap.Error(err))
		http.Error(w, `{"error":"failed to delete tenant"}`, http.StatusInternalServerError)
		return
	}

	go h.billing.DecrementTenantCount(context.Background(), instanceID)

	w.WriteHeader(http.StatusNoContent)
}

// Suspend handles POST /api/admin/tenants/:id/suspend.
func (h *TenantsHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err := h.service.SetActive(r.Context(), id, false); err != nil {
		h.log.Error("suspend tenant", zap.Error(err))
		http.Error(w, `{"error":"failed to suspend tenant"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unsuspend handles POST /api/admin/tenants/:id/unsuspend.
func (h *TenantsHandler) Unsuspend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err := h.service.SetActive(r.Context(), id, true); err != nil {
		h.log.Error("unsuspend tenant", zap.Error(err))
		http.Error(w, `{"error":"failed to unsuspend tenant"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RotateKey handles POST /api/admin/tenants/:id/rotate-key.
func (h *TenantsHandler) RotateKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

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
	if !h.checkPartnerOwns(r.Context(), id) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

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
