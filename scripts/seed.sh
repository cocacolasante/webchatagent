#!/usr/bin/env bash
# Provision Blueprint Automation's own chat widget demo tenant

set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080}"
ADMIN_KEY="${BLUEPRINT_ADMIN_KEY:-}"

if [ -z "$ADMIN_KEY" ]; then
  if [ -f .env ]; then
    ADMIN_KEY=$(grep BLUEPRINT_ADMIN_KEY .env | cut -d= -f2 | tr -d '"' | tr -d "'" | tr -d ' ')
  fi
fi

if [ -z "$ADMIN_KEY" ]; then
  echo "❌ BLUEPRINT_ADMIN_KEY not set. Set it in .env or environment."
  exit 1
fi

echo "🌱 Seeding Blueprint Automation demo tenant..."

RESPONSE=$(curl -s -X POST "${API_BASE}/api/admin/tenants" \
  -H "X-Blueprint-Admin-Key: ${ADMIN_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Blueprint Automation",
    "botName": "Aria",
    "primaryColor": "#6C63FF",
    "position": "bottom-right",
    "greeting": "Hi! I'\''m Aria, Blueprint Automation'\''s AI assistant. Ask me anything about our automation services, or I can help schedule a discovery call!",
    "plan": "pro",
    "businessInfo": {
      "name": "Blueprint Automation",
      "tagline": "AI-powered automation for small businesses",
      "industry": "Technology / SaaS",
      "website": "https://blueprintautomation.tech",
      "email": "hello@blueprintautomation.tech",
      "hours": "We respond within 24 hours. Automations run 24/7.",
      "services": [
        {"name": "AI Webchat Agent", "price": "$297 setup / $197/mo", "description": "AI chat widget for your website"},
        {"name": "Lead Pipeline", "price": "$497 setup / $149/mo", "description": "Automated lead capture to CRM"},
        {"name": "Email Triage", "price": "$299/mo", "description": "AI classifies and drafts email responses"},
        {"name": "Invoice Automation", "price": "$197 setup / $79/mo", "description": "Automated invoicing and payment follow-up"}
      ]
    },
    "knowledgeBase": {
      "faqs": [
        {
          "question": "What services do you offer?",
          "answer": "We offer AI Webchat Agents, Lead Capture Pipelines, Email Triage & Drafting, Invoice Automation, and Discord/Slack Command Centers for small businesses."
        },
        {
          "question": "How much does it cost?",
          "answer": "Our webchat agent starts at $297 setup with $197/month managed. Lead pipelines start at $497 setup. Email triage is $299/month. We also offer bundled packages. Book a call and we can find the right fit."
        },
        {
          "question": "How long does setup take?",
          "answer": "Most setups are complete within 48-72 hours. Some custom builds take up to a week."
        },
        {
          "question": "Do I need technical knowledge?",
          "answer": "No. We handle all the technical setup. You just fill in a short onboarding form and we take it from there."
        }
      ],
      "customInstructions": "Always be warm and helpful. Never quote exact custom pricing for complex projects — offer a discovery call instead. If someone asks about a service not listed, say we may be able to build it and offer to discuss."
    },
    "schedulerType": "calcom",
    "leadCaptureEnabled": true,
    "leadFormConfig": {
      "fields": ["name", "email", "phone"],
      "requiredFields": ["name", "email"],
      "triggerMessage": "I'\''d love to connect you with our team. Can I get your contact info?"
    }
  }')

echo ""
echo "✅ Tenant created!"
echo ""
echo "Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
echo ""

TENANT_ID=$(echo "$RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tenant',{}).get('id',''))" 2>/dev/null || echo "")
EMBED_CODE=$(echo "$RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('embedCode',''))" 2>/dev/null || echo "")

if [ -n "$EMBED_CODE" ]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📋 EMBED CODE (copy this to your website):"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "$EMBED_CODE"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

if [ -n "$TENANT_ID" ]; then
  echo ""
  echo "🎯 Tenant ID: $TENANT_ID"
  echo "🔗 Widget config: ${API_BASE}/widget-config?tid=${TENANT_ID}"
fi
