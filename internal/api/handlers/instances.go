package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// InstancesHandler handles product instance management.
type InstancesHandler struct {
	instanceSvc *tenant.InstanceService
	log         *zap.Logger
}

// NewInstancesHandler creates a new InstancesHandler.
func NewInstancesHandler(svc *tenant.InstanceService, log *zap.Logger) *InstancesHandler {
	return &InstancesHandler{instanceSvc: svc, log: log}
}

// Create handles POST /api/admin/instances.
func (h *InstancesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req tenant.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.ClientID == "" {
		http.Error(w, `{"error":"client_id is required"}`, http.StatusBadRequest)
		return
	}
	if req.InstanceNumber != 1 && req.InstanceNumber != 2 {
		http.Error(w, `{"error":"instance_number must be 1 or 2"}`, http.StatusBadRequest)
		return
	}

	inst, err := h.instanceSvc.Create(r.Context(), req)
	if err != nil {
		if err == tenant.ErrInstanceLimitReached {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "instance_limit_reached",
				"message": "This client already has 2 instances. Maximum of 2 instances allowed per product per client.",
			})
			return
		}
		h.log.Error("create instance", zap.Error(err))
		http.Error(w, `{"error":"failed to create instance"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(inst)
}

// List handles GET /api/admin/instances?client_id={uuid}.
func (h *InstancesHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(w, `{"error":"client_id query param is required"}`, http.StatusBadRequest)
		return
	}

	var instances []*tenant.ProductInstance
	var err error
	if partnerID, scoped := middleware.GetPartnerScope(r.Context()); scoped {
		instances, err = h.instanceSvc.ListByClientIDForPartner(r.Context(), clientID, partnerID)
	} else {
		instances, err = h.instanceSvc.ListByClientID(r.Context(), clientID)
	}
	if err != nil {
		h.log.Error("list instances", zap.Error(err))
		http.Error(w, `{"error":"failed to list instances"}`, http.StatusInternalServerError)
		return
	}

	totalTenants := 0
	for _, inst := range instances {
		totalTenants += inst.TenantCount
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tenant.InstancesResponse{
		Instances:    instances,
		TotalTenants: totalTenants,
		TotalMax:     len(instances) * tenant.MaxTenantsPerInstance,
	})
}
