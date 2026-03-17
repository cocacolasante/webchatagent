# Deployment Guide — Blueprint Chat

Fresh Ubuntu 22.04 → live Blueprint Chat in production.

## Prerequisites

A server with:
- Ubuntu 22.04 LTS
- 2 vCPU, 2GB RAM minimum
- Domain pointing to server IP (A record)
- Ports 80, 443, 8080 open in firewall

## Step 1: Install Docker & Git

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo apt-get install -y docker-compose-plugin

# Install Git
sudo apt-get install -y git

# Verify
docker --version
git --version
```

## Step 2: Clone & Configure

```bash
git clone https://github.com/blueprintautomation/blueprint-chat.git
cd blueprint-chat

# Copy and edit environment
cp .env.example .env
nano .env
```

Fill in all required values:
- `ANTHROPIC_API_KEY` — from console.anthropic.com
- `BLUEPRINT_ADMIN_KEY` — run: `openssl rand -hex 32`
- `ENCRYPTION_KEY` — run: `openssl rand -hex 32`
- `DATABASE_URL` — update username/password to match `POSTGRES_USER`/`POSTGRES_PASSWORD`

## Step 3: Build

```bash
# Install Node.js (for widget build)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Build everything
make build-widget build-admin build-all
```

## Step 4: Start Services

```bash
docker-compose up -d

# Verify all services are healthy
docker-compose ps
```

Expected: 6 containers running — api, admin, postgres, redis, n8n, nginx.

## Step 5: Run Migrations

```bash
# Wait ~10s for postgres to be fully ready, then:
docker-compose exec blueprint-chat-api ./blueprint-chat migrate up

# Or use the Makefile target:
DATABASE_URL="postgres://blueprint:changeme@localhost:5432/blueprint_chat?sslmode=disable" \
  migrate -path internal/db/migrations -database $DATABASE_URL up
```

## Step 6: Seed Demo Tenant

```bash
make seed
```

Copy the embed code from the output. You'll need it for Step 10.

## Step 7: Configure DNS

In your DNS provider, add an A record:
```
chat.blueprintautomation.tech → YOUR_SERVER_IP
```

Wait for propagation (1-5 minutes with low TTL).

## Step 8: SSL with Let's Encrypt

```bash
sudo apt-get install -y certbot

# Stop nginx temporarily
docker-compose stop nginx

# Get certificate
sudo certbot certonly --standalone -d chat.blueprintautomation.tech

# Update nginx.conf to use HTTPS
# Add SSL config block (see nginx/nginx.conf — uncomment SSL section)

# Restart
docker-compose up -d nginx
```

## Step 9: Import n8n Workflows

```bash
# n8n should be running at http://YOUR_IP:5678
make import-n8n
```

Log into n8n, configure credentials (Discord webhook, SMTP), and activate workflows.

## Step 10: Test

```bash
# 1. Check health
curl https://chat.blueprintautomation.tech/api/health

# 2. Create a test.html file
cat > /tmp/test.html << 'EOF'
<!DOCTYPE html>
<html>
<head><title>Widget Test</title></head>
<body>
<h1>Widget Test Page</h1>
PASTE_EMBED_CODE_HERE
</body>
</html>
EOF

# 3. Open test.html in browser and verify:
#    - Chat bubble appears
#    - Clicking opens panel
#    - Sending a message streams a response
```

## Step 11: Add Embed Code to blueprintautomation.tech

Paste the Blueprint Automation tenant embed code into your website's `</body>` tag.

## Updating (Zero Downtime)

```bash
git pull
make build-all
docker-compose up -d --build blueprint-chat-api
docker-compose up -d --build blueprint-chat-admin
```

## Monitoring

```bash
# Live logs
make logs

# Check container status
docker-compose ps

# API health
curl http://localhost:8080/api/health
```
