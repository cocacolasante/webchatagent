-- Create product_instances table for local instance tracking (up to 2 instances per client per product)
CREATE TABLE IF NOT EXISTS product_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       TEXT,
    instance_number INTEGER NOT NULL DEFAULT 1 CHECK (instance_number IN (1, 2)),
    portals_instance_id TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_instances_client_instance
    ON product_instances (client_id, instance_number)
    WHERE client_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_instances_portals
    ON product_instances (portals_instance_id)
    WHERE portals_instance_id IS NOT NULL;

-- Add product_instance_id to tenants
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS product_instance_id UUID REFERENCES product_instances(id);

CREATE INDEX IF NOT EXISTS idx_tenants_product_instance
    ON tenants (product_instance_id);

-- Migrate existing tenants: create one product_instance per portals_instance_id group,
-- then assign tenants to that instance.
WITH grouped AS (
    SELECT DISTINCT ON (portals_instance_id)
        portals_instance_id,
        client_id
    FROM tenants
    WHERE portals_instance_id IS NOT NULL
),
inserted AS (
    INSERT INTO product_instances (id, client_id, instance_number, portals_instance_id)
    SELECT gen_random_uuid(), g.client_id, 1, g.portals_instance_id
    FROM grouped g
    ON CONFLICT DO NOTHING
    RETURNING id, portals_instance_id
)
UPDATE tenants t
SET product_instance_id = i.id
FROM inserted i
WHERE t.portals_instance_id = i.portals_instance_id
  AND t.product_instance_id IS NULL;

-- For any remaining tenants without portals_instance_id, create individual instances
DO $$
DECLARE
    r RECORD;
    new_inst_id UUID;
BEGIN
    FOR r IN SELECT id FROM tenants WHERE product_instance_id IS NULL LOOP
        INSERT INTO product_instances (id, instance_number)
        VALUES (gen_random_uuid(), 1)
        RETURNING id INTO new_inst_id;

        UPDATE tenants SET product_instance_id = new_inst_id WHERE id = r.id;
    END LOOP;
END $$;
