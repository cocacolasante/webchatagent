# Blueprint Chat

AI-powered webchat agent for small businesses. Embeds on any website with one script tag.

## Features

- **One-line embed** — single `<script>` tag on any website
- **Multi-tenant** — one server, unlimited clients
- **Streaming AI** — real-time Claude responses via SSE
- **Lead capture** — inline form, webhook dispatch, email/Discord notifications
- **Appointment booking** — Cal.com, Calendly, Google Calendar support
- **Plugin architecture** — extensible without touching core code
- **Admin dashboard** — React + Tailwind, full tenant management

## Quick Start

```bash
# 1. Clone
git clone https://github.com/blueprintautomation/blueprint-chat.git
cd blueprint-chat

# 2. Configure
cp .env.example .env
# Edit .env — set ANTHROPIC_API_KEY, BLUEPRINT_ADMIN_KEY, ENCRYPTION_KEY, DATABASE_URL

# 3. Build
make build-widget build-admin build

# 4. Start
docker-compose up -d

# 5. Migrate
make migrate

# 6. Seed demo tenant
make seed
```

Visit `http://localhost:8080/api/health` to verify the server is running.
Admin dashboard at `http://localhost:3001`.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22, chi router |
| Database | PostgreSQL 16 |
| Cache/Sessions | Redis 7 |
| AI | Anthropic Claude (claude-sonnet-4-20250514) |
| Widget | Vanilla TypeScript, esbuild |
| Admin | React 18, Vite, TailwindCSS |
| Automation | n8n |
| Deployment | Docker, nginx |

## Project Structure

```
cmd/server/         — entry point
internal/
  api/              — HTTP handlers, middleware, router
  chat/             — SSE engine, session management
  knowledge/        — prompt assembly, document chunking
  tenant/           — multi-tenant CRUD + provisioning
  leads/            — lead capture + notifications
  scheduler/        — Cal.com, Calendly, Google Calendar adapters
  plugins/          — plugin registry + example plugin
  notification/     — email + Discord
  db/               — PostgreSQL + Redis + migrations
  config/           — environment configuration
pkg/encrypt/        — AES-256-GCM encryption
widget/             — Vanilla TS chat widget
admin/              — React admin dashboard
n8n-workflows/      — automation workflow JSON files
docs/               — architecture decisions, deployment guide
```

## Documentation

- [Architecture Decisions](docs/ARCHITECTURE_DECISIONS.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Plugin Interface](docs/PLUGIN_INTERFACE.md)
- [Scheduler Interface](docs/SCHEDULER_INTERFACE.md)
- [Client Handoff Guide](docs/CLIENT_HANDOFF.md)
- [n8n Workflows](n8n-workflows/README.md)

## API Reference

### Public Endpoints

```
GET  /api/health              — System health check
GET  /widget.js               — Widget bundle
GET  /widget-config?tid=ID    — Widget configuration
POST /api/chat/stream?tid=ID  — SSE chat stream
POST /api/leads?tid=ID        — Lead capture
POST /api/booking/availability?tid=ID  — Available slots
POST /api/booking/create?tid=ID        — Create booking
```

### Admin Endpoints (require X-Blueprint-Admin-Key header)

```
POST   /api/admin/tenants              — Create tenant
GET    /api/admin/tenants              — List tenants
GET    /api/admin/tenants/:id          — Get tenant
PUT    /api/admin/tenants/:id          — Update tenant
DELETE /api/admin/tenants/:id          — Delete tenant
POST   /api/admin/tenants/:id/rotate-key  — Rotate API key
POST   /api/admin/tenants/:id/knowledge   — Update knowledge base
GET    /api/admin/tenants/:id/leads       — List leads
```

---

## Blueprint Command Integration

This product supports the [Blueprint Command](../portals/) partner/reseller portal. Partners can provision and manage tenants through a shared admin API.

### Database Changes

Three columns were added to the `tenants` table via migration:

| Column | Type | Description |
|--------|------|-------------|
| `partner_id` | `UUID` (nullable) | ID of the reseller partner who owns this tenant |
| `managed_by` | `TEXT` (default `'bpa'`) | Who manages this tenant (`'bpa'` or `'partner'`) |
| `client_id` | `TEXT` (nullable) | External CRM or client reference ID |

### API Key Types

| Key Prefix | Access Level |
|------------|-------------|
| `bpa_super_{random}` | Super admin — full access across all tenants |
| `bpa_partner_{uuid}_{random}` | Partner — scoped to their own tenants only |
| (regular key) | Standard admin key validated against `BLUEPRINT_ADMIN_KEY` env var |

### Partner-Scoped Endpoints

All admin endpoints respect partner scoping. When a `bpa_partner_` key is used, list endpoints automatically filter to only return that partner's tenants.

### Stats Endpoint

```
GET /api/admin/tenants/:id/stats
```

Called by Blueprint Command to populate client portal usage dashboards.

**Example response:**
```json
{
  "product_key": "webchatagent",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "period": "last_30_days",
  "stats": {
    "conversations_total": 142,
    "messages_total": 891,
    "leads_captured": 23
  },
  "summary": "Your AI chat agent handled 142 conversations and captured 23 leads this month."
}
```

### Applying Migrations

```bash
./scripts/apply-reseller-migrations.sh
```

---

Built by [Blueprint Automation](https://blueprintautomation.tech)
