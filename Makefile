.PHONY: up down logs build run migrate seed build-widget build-admin build-all import-n8n provision test lint clean

# ── Docker ──────────────────────────────────────────────────────────────────────

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# ── Go ──────────────────────────────────────────────────────────────────────────

build:
	go build -ldflags="-s -w" -o bin/blueprint-chat ./cmd/server
	@echo "✅ Go binary built: bin/blueprint-chat"

run:
	go run ./cmd/server

# ── Database ─────────────────────────────────────────────────────────────────────

migrate:
	@echo "Running migrations..."
	docker-compose exec blueprint-chat-api ./blueprint-chat migrate up 2>/dev/null || \
	  migrate -path internal/db/migrations -database "$(DATABASE_URL)" up
	@echo "✅ Migrations complete"

migrate-down:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" down 1

migrate-status:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" version

seed:
	./scripts/seed.sh

# ── Widget ──────────────────────────────────────────────────────────────────────

build-widget:
	cd widget && npm install && npm run build
	@SIZE=$$(wc -c < widget/dist/widget.js | tr -d ' '); \
	  MAX=51200; \
	  if [ "$$SIZE" -gt "$$MAX" ]; then \
	    echo "❌ Widget too large: $$SIZE bytes (max 50KB raw)"; exit 1; \
	  else \
	    echo "✅ Widget size OK: $$SIZE bytes"; \
	  fi

# ── Admin ──────────────────────────────────────────────────────────────────────

build-admin:
	cd admin && npm install && npm run build
	@echo "✅ Admin dashboard built"

# ── Full build ──────────────────────────────────────────────────────────────────

build-all: build-widget build-admin build
	@echo "✅ Full build complete"

# ── n8n ─────────────────────────────────────────────────────────────────────────

import-n8n:
	./scripts/import-n8n-workflows.sh

# ── Provisioning ────────────────────────────────────────────────────────────────

provision:
	./scripts/provision-tenant.sh

# ── Testing & Quality ────────────────────────────────────────────────────────────

test:
	go test ./... -v -race -timeout 60s

test-short:
	go test ./... -short

lint:
	golangci-lint run ./...

vet:
	go vet ./...

# ── Cleanup ──────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ widget/dist/ admin/dist/
	@echo "✅ Clean complete"
