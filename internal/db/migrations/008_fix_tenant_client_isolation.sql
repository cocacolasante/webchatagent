-- 008_fix_tenant_client_isolation.sql

-- Index to support ListByClientID queries
CREATE INDEX IF NOT EXISTS idx_tenants_client_id
    ON tenants(client_id)
    WHERE client_id IS NOT NULL;

-- Back-fill client_id on product_instances that are NULL
-- by reading the client_id from their associated tenants
UPDATE product_instances pi
SET client_id = sub.client_id
FROM (
    SELECT DISTINCT ON (product_instance_id)
        product_instance_id,
        client_id
    FROM tenants
    WHERE client_id IS NOT NULL
      AND product_instance_id IS NOT NULL
    ORDER BY product_instance_id, created_at ASC
) sub
WHERE pi.id = sub.product_instance_id
  AND pi.client_id IS NULL;
