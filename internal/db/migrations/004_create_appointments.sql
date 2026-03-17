CREATE TABLE IF NOT EXISTS appointments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id   UUID REFERENCES conversations(id),
    lead_id           UUID REFERENCES leads(id),
    scheduler_type    TEXT NOT NULL,
    external_id       TEXT NOT NULL,
    attendee_email    TEXT NOT NULL,
    attendee_name     TEXT NOT NULL,
    start_time        TIMESTAMPTZ NOT NULL,
    end_time          TIMESTAMPTZ NOT NULL,
    meeting_url       TEXT,
    confirm_url       TEXT,
    cancel_url        TEXT,
    status            TEXT NOT NULL DEFAULT 'confirmed',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_appointments_tenant ON appointments(tenant_id);
