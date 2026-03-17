#!/usr/bin/env bash
# Import all n8n workflow JSON files via the n8n API

set -euo pipefail

N8N_BASE="${N8N_BASE_URL:-http://localhost:5678}"
N8N_KEY="${N8N_API_KEY:-}"
N8N_USER="${N8N_USER:-admin}"
N8N_PASS="${N8N_PASSWORD:-changeme}"

WORKFLOW_DIR="./n8n-workflows"

if [ ! -d "$WORKFLOW_DIR" ]; then
  echo "❌ Workflow directory not found: $WORKFLOW_DIR"
  exit 1
fi

AUTH_HEADER=""
if [ -n "$N8N_KEY" ]; then
  AUTH_HEADER="X-N8N-API-KEY: $N8N_KEY"
else
  AUTH_HEADER="Authorization: Basic $(echo -n "${N8N_USER}:${N8N_PASS}" | base64)"
fi

echo "📦 Importing n8n workflows from $WORKFLOW_DIR..."

for file in "$WORKFLOW_DIR"/*.json; do
  name=$(basename "$file" .json)
  echo "  → Importing: $name"

  curl -s -X POST "${N8N_BASE}/api/v1/workflows" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d @"$file" > /dev/null && echo "    ✅ $name" || echo "    ❌ Failed: $name"
done

echo ""
echo "✅ Workflow import complete. Visit ${N8N_BASE} to activate them."
