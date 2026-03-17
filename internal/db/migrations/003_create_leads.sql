CREATE TABLE IF NOT EXISTS leads (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id),
    first_name      TEXT NOT NULL,
    last_name       TEXT,
    email           TEXT NOT NULL,
    phone           TEXT,
    source_url      TEXT,
    utm_source      TEXT,
    utm_medium      TEXT,
    utm_campaign    TEXT,
    session_summary TEXT,
    status          TEXT NOT NULL DEFAULT 'new',
    webhook_sent    BOOLEAN NOT NULL DEFAULT false,
    webhook_sent_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_leads_dedup ON leads(tenant_id, email, date_trunc('day', created_at));
CREATE INDEX IF NOT EXISTS idx_leads_tenant ON leads(tenant_id);
