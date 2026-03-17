# Architecture Decisions — Blueprint Chat

## Language: Go

**Decision:** Go with `chi` router.

**Why Go over Python:**
- Go's goroutine model handles hundreds of simultaneous SSE streams with minimal memory overhead. Each SSE connection is a long-lived HTTP connection — Go's goroutine-per-connection model is ideal.
- Single binary deployment: `go build` → one executable. To patch all client servers: copy binary, restart process. No virtualenvs, no pip freeze, no dependency hell.
- Strong typing catches bugs at compile time. The multi-tenant middleware pattern is clear and idiomatic in Go.
- `net/http` + `chi` give us full control over SSE flushing without framework abstractions getting in the way.

**Why not Python/FastAPI:**
- FastAPI's async streaming works but requires careful async context management.
- Python deployment requires managing interpreter versions, virtual environments, and dependency chains across many client servers.
- Memory overhead per connection is higher in Python async than Go goroutines.

## Router: chi

**Decision:** `github.com/go-chi/chi/v5`

**Why chi over Gin or Echo:**
- chi uses standard `net/http` — handlers are `http.HandlerFunc`, not framework-specific types. This means future migration is trivial and middleware from the stdlib ecosystem works directly.
- Lightweight: chi adds ~5KB to the binary vs Gin/Echo adding more.
- Composable middleware groups make the tenant/auth/ratelimit chain clean.

## Database Driver: pgx/v5

**Decision:** `github.com/jackc/pgx/v5`

**Why pgx over database/sql:**
- pgx has native PostgreSQL type support — JSONB, UUID, timestamptz all handled correctly without manual conversion.
- `pgxpool` provides connection pooling built in with health checks.
- Faster than database/sql for PostgreSQL-specific queries.

## Session Storage: Redis

**Decision:** Redis via `go-redis/v9`

**Why Redis for sessions:**
- Sessions are ephemeral (30-minute TTL) — Redis's TTL-based expiry is a perfect fit.
- Sub-millisecond reads for session lookups on every chat message.
- Same Redis instance serves as rate limit counter and prompt cache.

## Widget: Vanilla TypeScript (no framework)

**Decision:** Single IIFE compiled with esbuild, no React.

**Why no React:**
- The widget embeds on any website including those running different React versions, Vue, Angular, or no framework.
- React 18 + React DOM adds ~45KB gzipped. Our entire widget is under 50KB gzipped.
- Shadow DOM was considered but skipped due to compatibility issues with older browsers and some website builders.
- CSS class prefixing (`bp-chat-`) prevents all conflicts.

## SSE over WebSockets

**Decision:** Server-Sent Events for chat streaming.

**Why SSE over WebSockets:**
- SSE is unidirectional (server → client), which is exactly what streaming AI responses need. Chat messages go server → client; user messages are a standard POST.
- SSE works over HTTP/1.1 and HTTP/2 without protocol upgrade.
- Simpler reconnection logic: browsers handle SSE reconnect automatically.
- No WebSocket upgrade overhead per connection.

## Plugin System

**Decision:** Interface-based plugin registry with lifecycle hooks.

**Why this design:**
- Plugins implement a single `Plugin` interface — no magic, no reflection.
- Hooks are synchronous from the plugin's perspective but the registry fires them all. Plugins failing don't block the main request.
- Future: move to dynamic plugin loading (`.so` files or subprocess plugins) without changing the interface.

## Encryption: AES-256-GCM

**Decision:** AES-256-GCM via `golang.org/x/crypto` for scheduler API keys.

**Why:**
- GCM provides authenticated encryption — tampering is detectable.
- Keys stored in environment variables (not in DB), ciphertext in DB.
- Each encryption call generates a unique nonce, stored with the ciphertext.

## Multi-tenancy: Row-level with middleware resolution

**Decision:** All tenants in shared tables, tenant ID resolved per-request via middleware.

**Why shared tables over separate schemas/databases:**
- Simpler migrations (one schema to maintain).
- Easier cross-tenant admin queries.
- At the scale of hundreds of small business clients, row-level is sufficient.
- Tenant isolation is enforced at the application layer — every query includes `WHERE tenant_id = $1`.

## Deferred / Not Built

- Vector search for knowledge base (future: pgvector)
- Voice input/output plugin
- Live handoff plugin
- Payment collection plugin
- WebSocket support (not needed — SSE covers the use case)
- Tenant subdomain routing (stubbed in middleware, not fully implemented)
