package tenant

import "time"

const (
	MaxTenantsPerInstance    = 15
	IncludedSeatsPerInstance = 3
	MaxInstancesPerClient    = 2
)

// ProductInstance represents a provisioned product deployment for a client.
type ProductInstance struct {
	ID                string    `json:"instanceId"`
	ClientID          string    `json:"clientId,omitempty"`
	InstanceNumber    int       `json:"instanceNumber"`
	PortalsInstanceID string    `json:"portalsInstanceId,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`

	// Computed at query time
	TenantCount   int  `json:"tenantCount"`
	MaxTenants    int  `json:"maxTenants"`
	IncludedSeats int  `json:"includedSeats"`
	OverageSeats  int  `json:"overageSeats"`
	AtCapacity    bool `json:"atCapacity"`
}

// CreateInstanceRequest is the payload for POST /api/admin/instances.
type CreateInstanceRequest struct {
	ClientID       string `json:"client_id"`
	InstanceNumber int    `json:"instance_number"`
}

// InstancesResponse is the payload for GET /api/admin/instances.
type InstancesResponse struct {
	Instances    []*ProductInstance `json:"instances"`
	TotalTenants int                `json:"total_tenants"`
	TotalMax     int                `json:"total_max"`
}
