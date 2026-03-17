CREATE TABLE IF NOT EXISTS tenants (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key          TEXT NOT NULL UNIQUE,
    name             TEXT NOT NULL,
    plan             TEXT NOT NULL DEFAULT 'starter',
    is_active        BOOLEAN NOT NULL DEFAULT true,

    -- Branding
    bot_name         TEXT NOT NULL DEFAULT 'Assistant',
    avatar_url       TEXT,
    primary_color    TEXT NOT NULL DEFAULT '#6C63FF',
    greeting         TEXT NOT NULL DEFAULT 'Hi! How can I help you today?',
    position         TEXT NOT NULL DEFAULT 'bottom-right',

    -- Business context (stored as JSONB for flexibility)
    business_info    JSONB NOT NULL DEFAULT '{}',
    knowledge_base   JSONB NOT NULL DEFAULT '{}',

    -- Scheduler
    scheduler_type   TEXT,
    scheduler_config JSONB,

    -- Lead capture
    lead_capture_enabled  BOOLEAN NOT NULL DEFAULT true,
    lead_form_config      JSONB NOT NULL DEFAULT '{}',
    lead_webhook_url      TEXT,
    lead_notify_email     TEXT,

    -- Notifications
    discord_webhook_url   TEXT,

    -- Rate limits
    max_messages_per_day  INTEGER NOT NULL DEFAULT 1000,
    max_messages_per_min  INTEGER NOT NULL DEFAULT 30,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenants_api_key ON tenants(api_key);
