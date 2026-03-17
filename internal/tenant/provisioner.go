package tenant

import (
	"context"
	"fmt"
)

// ProvisionResult contains everything a new tenant needs to get started.
type ProvisionResult struct {
	Tenant    *Tenant `json:"tenant"`
	EmbedCode string  `json:"embedCode"`
}

// Provisioner handles full APS tenant provisioning.
type Provisioner struct {
	service *Service
	baseURL string
}

// NewProvisioner creates a new Provisioner.
func NewProvisioner(service *Service, baseURL string) *Provisioner {
	return &Provisioner{service: service, baseURL: baseURL}
}

// Provision creates a new tenant and returns provisioning details including the embed code.
func (p *Provisioner) Provision(ctx context.Context, req CreateTenantRequest) (*ProvisionResult, error) {
	t, err := p.service.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	embedCode := p.generateEmbedCode(t)

	return &ProvisionResult{
		Tenant:    t,
		EmbedCode: embedCode,
	}, nil
}

// generateEmbedCode produces the one-line script tag embed snippet.
func (p *Provisioner) generateEmbedCode(t *Tenant) string {
	return fmt.Sprintf(
		`<!-- Blueprint Chat — powered by Blueprint Automation -->`+"\n"+
			`<script src="%s/widget.js" data-tenant-id="%s" data-position="%s" async></script>`,
		p.baseURL, t.ID, t.Position,
	)
}
