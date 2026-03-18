package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// TenantLimitClient calls Blueprint Command (portals) to enforce per-tenant billing limits.
// All methods fail open: if portals is unreachable or returns an error, tenant creation proceeds.
type TenantLimitClient struct {
	portalsURL  string
	internalKey string
	enabled     bool
	httpClient  *http.Client
}

// CheckResult is returned by CheckTenantAllowed.
type CheckResult struct {
	Allowed bool `json:"allowed"`
}

// NewTenantLimitClient creates a billing client. If portalsURL is empty, billing is disabled.
func NewTenantLimitClient(portalsURL, internalKey string) *TenantLimitClient {
	return &TenantLimitClient{
		portalsURL:  portalsURL,
		internalKey: internalKey,
		enabled:     portalsURL != "" && internalKey != "",
		httpClient:  &http.Client{Timeout: 3 * time.Second},
	}
}

// CheckTenantAllowed returns true if a new tenant may be created for this portals instance.
// Fails open (returns allowed=true) on any error.
func (c *TenantLimitClient) CheckTenantAllowed(ctx context.Context, instanceID string) bool {
	if !c.enabled || instanceID == "" {
		return true
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.portalsURL+"/api/internal/tenant-check?instance_id="+instanceID, nil)
	if err != nil {
		return true
	}
	req.Header.Set("X-BPA-Internal-Key", c.internalKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return true // fail open
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return true // fail open
	}
	var result CheckResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true
	}
	return result.Allowed
}

// IncrementTenantCount notifies portals that a tenant was created. Fire-and-forget.
func (c *TenantLimitClient) IncrementTenantCount(ctx context.Context, instanceID string) {
	if !c.enabled || instanceID == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{"instance_id": instanceID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.portalsURL+"/api/internal/tenant-increment", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("X-BPA-Internal-Key", c.internalKey)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := c.httpClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
}

// DecrementTenantCount notifies portals that a tenant was deleted. Fire-and-forget.
func (c *TenantLimitClient) DecrementTenantCount(ctx context.Context, instanceID string) {
	if !c.enabled || instanceID == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{"instance_id": instanceID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.portalsURL+"/api/internal/tenant-decrement", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("X-BPA-Internal-Key", c.internalKey)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := c.httpClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
}
