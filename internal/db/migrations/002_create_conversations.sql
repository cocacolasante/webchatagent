CREATE TABLE IF NOT EXISTS conversations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    session_id   TEXT NOT NULL,
    source_url   TEXT,
    utm_params   JSONB,
    lead_id      UUID,
    appointment_id TEXT,
    message_count INTEGER NOT NULL DEFAULT 0,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_tenant ON conversations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_conversations_session ON conversations(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
