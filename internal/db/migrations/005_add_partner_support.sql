-- 005_add_partner_support.sql
-- Blueprint Command reseller retrofit

ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS partner_id    UUID        DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS managed_by    TEXT        NOT NULL DEFAULT 'bpa',
    ADD COLUMN IF NOT EXISTS client_id     TEXT        DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_tenants_partner_id
    ON tenants(partner_id)
    WHERE partner_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tenants_client_id
    ON tenants(client_id)
    WHERE client_id IS NOT NULL;

COMMENT ON COLUMN tenants.partner_id IS 'UUID of the Blueprint Command partner that manages this tenant. NULL = BPA direct.';
COMMENT ON COLUMN tenants.managed_by IS 'Who manages this tenant: bpa (Blueprint direct) or partner (reseller).';
COMMENT ON COLUMN tenants.client_id IS 'UUID string of the client record in blueprint_command database.';
