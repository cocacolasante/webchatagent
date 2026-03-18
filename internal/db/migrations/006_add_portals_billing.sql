-- Add portals_instance_id to tenants for per-tenant billing via Blueprint Command
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS portals_instance_id TEXT;
CREATE INDEX IF NOT EXISTS idx_tenants_portals_instance_id ON tenants(portals_instance_id) WHERE portals_instance_id IS NOT NULL;
