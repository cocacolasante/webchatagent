package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInstanceLimitReached = errors.New("client already has 2 instances for this product")
var ErrInvalidInstanceNumber = errors.New("instance_number must be 1 or 2")

// InstanceService manages product instances.
type InstanceService struct {
	db *pgxpool.Pool
}

// NewInstanceService creates a new InstanceService.
func NewInstanceService(db *pgxpool.Pool) *InstanceService {
	return &InstanceService{db: db}
}

// CountByInstance returns the number of tenants in a given product instance.
func (s *InstanceService) CountByInstance(ctx context.Context, instanceID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenants WHERE product_instance_id = $1`, instanceID,
	).Scan(&count)
	return count, err
}

// GetByID returns a product instance with computed tenant counts.
func (s *InstanceService) GetByID(ctx context.Context, id string) (*ProductInstance, error) {
	inst := &ProductInstance{}
	err := s.db.QueryRow(ctx, `
		SELECT pi.id, COALESCE(pi.client_id,''), pi.instance_number,
		       COALESCE(pi.portals_instance_id,''), pi.created_at,
		       COUNT(t.id) AS tenant_count
		FROM product_instances pi
		LEFT JOIN tenants t ON t.product_instance_id = pi.id
		WHERE pi.id = $1
		GROUP BY pi.id, pi.client_id, pi.instance_number, pi.portals_instance_id, pi.created_at
	`, id).Scan(
		&inst.ID, &inst.ClientID, &inst.InstanceNumber,
		&inst.PortalsInstanceID, &inst.CreatedAt, &inst.TenantCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get instance: %w", err)
	}
	inst.MaxTenants = MaxTenantsPerInstance
	inst.IncludedSeats = IncludedSeatsPerInstance
	inst.OverageSeats = max(0, inst.TenantCount-IncludedSeatsPerInstance)
	inst.AtCapacity = inst.TenantCount >= MaxTenantsPerInstance
	return inst, nil
}

// ListByClientID returns all product instances for a client.
func (s *InstanceService) ListByClientID(ctx context.Context, clientID string) ([]*ProductInstance, error) {
	rows, err := s.db.Query(ctx, `
		SELECT pi.id, COALESCE(pi.client_id,''), pi.instance_number,
		       COALESCE(pi.portals_instance_id,''), pi.created_at,
		       COUNT(t.id) AS tenant_count
		FROM product_instances pi
		LEFT JOIN tenants t ON t.product_instance_id = pi.id
		WHERE pi.client_id = $1
		GROUP BY pi.id, pi.client_id, pi.instance_number, pi.portals_instance_id, pi.created_at
		ORDER BY pi.instance_number ASC
	`, clientID)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	var instances []*ProductInstance
	for rows.Next() {
		inst := &ProductInstance{}
		if err := rows.Scan(
			&inst.ID, &inst.ClientID, &inst.InstanceNumber,
			&inst.PortalsInstanceID, &inst.CreatedAt, &inst.TenantCount,
		); err != nil {
			return nil, err
		}
		inst.MaxTenants = MaxTenantsPerInstance
		inst.IncludedSeats = IncludedSeatsPerInstance
		inst.OverageSeats = max(0, inst.TenantCount-IncludedSeatsPerInstance)
		inst.AtCapacity = inst.TenantCount >= MaxTenantsPerInstance
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// ListByClientIDForPartner returns instances for a client only if they contain
// tenants belonging to the given partner, ensuring cross-partner isolation.
func (s *InstanceService) ListByClientIDForPartner(ctx context.Context, clientID, partnerID string) ([]*ProductInstance, error) {
	rows, err := s.db.Query(ctx, `
		SELECT pi.id, COALESCE(pi.client_id,''), pi.instance_number,
		       COALESCE(pi.portals_instance_id,''), pi.created_at,
		       COUNT(t.id) AS tenant_count
		FROM product_instances pi
		INNER JOIN tenants t ON t.product_instance_id = pi.id AND t.partner_id = $2
		WHERE pi.client_id = $1
		GROUP BY pi.id, pi.client_id, pi.instance_number, pi.portals_instance_id, pi.created_at
		ORDER BY pi.instance_number ASC
	`, clientID, partnerID)
	if err != nil {
		return nil, fmt.Errorf("list instances for partner: %w", err)
	}
	defer rows.Close()

	var instances []*ProductInstance
	for rows.Next() {
		inst := &ProductInstance{}
		if err := rows.Scan(
			&inst.ID, &inst.ClientID, &inst.InstanceNumber,
			&inst.PortalsInstanceID, &inst.CreatedAt, &inst.TenantCount,
		); err != nil {
			return nil, err
		}
		inst.MaxTenants = MaxTenantsPerInstance
		inst.IncludedSeats = IncludedSeatsPerInstance
		inst.OverageSeats = max(0, inst.TenantCount-IncludedSeatsPerInstance)
		inst.AtCapacity = inst.TenantCount >= MaxTenantsPerInstance
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// Create creates a new product instance for a client.
// Returns ErrInstanceLimitReached if the client already has 2 instances.
func (s *InstanceService) Create(ctx context.Context, req CreateInstanceRequest) (*ProductInstance, error) {
	if req.InstanceNumber != 1 && req.InstanceNumber != 2 {
		return nil, ErrInvalidInstanceNumber
	}

	// Count existing instances for this client
	var existing int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_instances WHERE client_id = $1`, req.ClientID,
	).Scan(&existing)
	if err != nil {
		return nil, fmt.Errorf("count instances: %w", err)
	}
	if existing >= MaxInstancesPerClient {
		return nil, ErrInstanceLimitReached
	}

	var id string
	err = s.db.QueryRow(ctx, `
		INSERT INTO product_instances (id, client_id, instance_number)
		VALUES (gen_random_uuid(), $1, $2)
		RETURNING id
	`, req.ClientID, req.InstanceNumber).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create instance: %w", err)
	}

	return s.GetByID(ctx, id)
}

// EnsureDefault creates or returns instance 1 for a portals_instance_id.
// Called at tenant creation time when no product_instance_id is provided.
func (s *InstanceService) EnsureDefault(ctx context.Context, clientID, portalsInstanceID string) (*ProductInstance, error) {
	// Try by portals_instance_id first
	if portalsInstanceID != "" {
		var id string
		err := s.db.QueryRow(ctx,
			`SELECT id FROM product_instances WHERE portals_instance_id = $1`, portalsInstanceID,
		).Scan(&id)
		if err == nil {
			return s.GetByID(ctx, id)
		}
	}

	// Try by client_id + instance_number=1
	if clientID != "" {
		var id string
		err := s.db.QueryRow(ctx,
			`SELECT id FROM product_instances WHERE client_id = $1 AND instance_number = 1`, clientID,
		).Scan(&id)
		if err == nil {
			return s.GetByID(ctx, id)
		}
	}

	// Create a new instance 1
	var id string
	var cid *string
	if clientID != "" {
		cid = &clientID
	}
	var pid *string
	if portalsInstanceID != "" {
		pid = &portalsInstanceID
	}
	err := s.db.QueryRow(ctx, `
		INSERT INTO product_instances (id, client_id, instance_number, portals_instance_id)
		VALUES (gen_random_uuid(), $1, 1, $2)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, cid, pid).Scan(&id)
	if err != nil || id == "" {
		// Race condition: another request created it, try fetching again
		if portalsInstanceID != "" {
			s.db.QueryRow(ctx,
				`SELECT id FROM product_instances WHERE portals_instance_id = $1`, portalsInstanceID,
			).Scan(&id)
		} else if clientID != "" {
			s.db.QueryRow(ctx,
				`SELECT id FROM product_instances WHERE client_id = $1 AND instance_number = 1`, clientID,
			).Scan(&id)
		}
	}
	if id == "" {
		return nil, fmt.Errorf("failed to ensure default instance")
	}
	return s.GetByID(ctx, id)
}

// ListAll returns all instances with tenant counts (for health endpoint).
func (s *InstanceService) ListAll(ctx context.Context) ([]*ProductInstance, error) {
	rows, err := s.db.Query(ctx, `
		SELECT pi.id, COALESCE(pi.client_id,''), pi.instance_number,
		       COALESCE(pi.portals_instance_id,''), pi.created_at,
		       COUNT(t.id) AS tenant_count
		FROM product_instances pi
		LEFT JOIN tenants t ON t.product_instance_id = pi.id
		GROUP BY pi.id, pi.client_id, pi.instance_number, pi.portals_instance_id, pi.created_at
		ORDER BY pi.instance_number ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*ProductInstance
	for rows.Next() {
		inst := &ProductInstance{}
		if err := rows.Scan(
			&inst.ID, &inst.ClientID, &inst.InstanceNumber,
			&inst.PortalsInstanceID, &inst.CreatedAt, &inst.TenantCount,
		); err != nil {
			return nil, err
		}
		inst.MaxTenants = MaxTenantsPerInstance
		inst.IncludedSeats = IncludedSeatsPerInstance
		inst.OverageSeats = max(0, inst.TenantCount-IncludedSeatsPerInstance)
		inst.AtCapacity = inst.TenantCount >= MaxTenantsPerInstance
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

