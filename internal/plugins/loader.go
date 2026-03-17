package plugins

// Loader discovers and loads plugins for a tenant.
// Future: load from config, support dynamic plugin loading.
type Loader struct {
	registry *Registry
}

// NewLoader creates a new plugin Loader.
func NewLoader(registry *Registry) *Loader {
	return &Loader{registry: registry}
}

// LoadForTenant returns enabled plugins for a specific tenant.
func (l *Loader) LoadForTenant(tenantID string) []Plugin {
	return l.registry.EnabledForTenant(tenantID)
}
