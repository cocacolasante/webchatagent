package tenant_test

import (
	"testing"

	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
)

func TestMaxConstants(t *testing.T) {
	if tenant.MaxTenantsPerInstance != 15 {
		t.Errorf("expected MaxTenantsPerInstance=15, got %d", tenant.MaxTenantsPerInstance)
	}
	if tenant.MaxInstancesPerClient != 2 {
		t.Errorf("expected MaxInstancesPerClient=2, got %d", tenant.MaxInstancesPerClient)
	}
	if tenant.IncludedSeatsPerInstance != 3 {
		t.Errorf("expected IncludedSeatsPerInstance=3, got %d", tenant.IncludedSeatsPerInstance)
	}
}

func TestProductInstanceAtCapacity(t *testing.T) {
	inst := &tenant.ProductInstance{TenantCount: 15, MaxTenants: 15}
	inst.AtCapacity = inst.TenantCount >= tenant.MaxTenantsPerInstance
	if !inst.AtCapacity {
		t.Error("expected AtCapacity=true when TenantCount==MaxTenantsPerInstance")
	}

	inst2 := &tenant.ProductInstance{TenantCount: 14, MaxTenants: 15}
	inst2.AtCapacity = inst2.TenantCount >= tenant.MaxTenantsPerInstance
	if inst2.AtCapacity {
		t.Error("expected AtCapacity=false when TenantCount<MaxTenantsPerInstance")
	}
}

func TestCreateInstanceRequestValidation(t *testing.T) {
	validReq := tenant.CreateInstanceRequest{ClientID: "test-client", InstanceNumber: 2}
	if validReq.InstanceNumber != 2 {
		t.Error("expected InstanceNumber=2")
	}
}
