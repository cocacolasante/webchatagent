#!/usr/bin/env bash
# CLI: provision a new tenant interactively

set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080}"
ADMIN_KEY="${BLUEPRINT_ADMIN_KEY:-}"

if [ -z "$ADMIN_KEY" ] && [ -f .env ]; then
  ADMIN_KEY=$(grep BLUEPRINT_ADMIN_KEY .env | cut -d= -f2 | tr -d '"' | tr -d "'" | tr -d ' ')
fi

if [ -z "$ADMIN_KEY" ]; then
  echo "❌ BLUEPRINT_ADMIN_KEY not set."
  exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Blueprint Chat — New Tenant"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

read -rp "Business name: " BUSINESS_NAME
read -rp "Bot name [Assistant]: " BOT_NAME
BOT_NAME="${BOT_NAME:-Assistant}"
read -rp "Primary color [#6C63FF]: " PRIMARY_COLOR
PRIMARY_COLOR="${PRIMARY_COLOR:-#6C63FF}"
read -rp "Greeting message: " GREETING
GREETING="${GREETING:-Hi! How can I help you today?}"

PAYLOAD=$(cat <<EOF
{
  "name": "${BUSINESS_NAME}",
  "botName": "${BOT_NAME}",
  "primaryColor": "${PRIMARY_COLOR}",
  "greeting": "${GREETING}",
  "leadCaptureEnabled": true,
  "leadFormConfig": {
    "fields": ["name", "email", "phone"],
    "requiredFields": ["name", "email"],
    "triggerMessage": "I would love to connect you with our team. Can I get your contact info?"
  }
}
EOF
)

echo ""
echo "⚙️  Creating tenant..."

RESPONSE=$(curl -s -X POST "${API_BASE}/api/admin/tenants" \
  -H "X-Blueprint-Admin-Key: ${ADMIN_KEY}" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD")

EMBED_CODE=$(echo "$RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('embedCode',''))" 2>/dev/null || echo "")
TENANT_ID=$(echo "$RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tenant',{}).get('id',''))" 2>/dev/null || echo "")
API_KEY=$(echo "$RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tenant',{}).get('apiKey',''))" 2>/dev/null || echo "")

echo ""
echo "✅ Tenant provisioned!"
echo ""
echo "Tenant ID : $TENANT_ID"
echo "API Key   : $API_KEY"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "EMBED CODE:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "$EMBED_CODE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
