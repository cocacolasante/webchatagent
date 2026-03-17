# Blueprint Chat — n8n Workflows

Four automation workflows to connect Blueprint Chat with your notification and CRM systems.

## Workflows

| File | Purpose | Trigger |
|------|---------|---------|
| `chat-lead-notification.json` | Notify team + forward to pipeline on lead capture | Webhook |
| `chat-booking-confirmation.json` | Notify team + update CRM on appointment booked | Webhook |
| `chat-tenant-provisioning.json` | Auto-provision tenant + send onboarding email (APS) | Webhook |
| `chat-daily-digest.json` | Daily system health + stats to Discord | Schedule (8am) |

## Import Guide

### Via n8n UI
1. Open n8n at `http://localhost:5678`
2. Go to **Workflows** → **Import from file**
3. Upload each `.json` file
4. Configure credentials (see below)
5. Activate each workflow

### Via Script
```bash
make import-n8n
```

## Required Credentials

Set these in n8n under **Credentials**:

### Discord Webhook
- Type: Discord Webhook
- Name: `Blueprint Discord Webhook`
- Webhook URL: Your Discord channel webhook URL

### SMTP (for provisioning emails)
- Type: SMTP
- Name: `Blueprint SMTP`
- Host, port, user, password from your email provider

### Environment Variables in n8n
Set these via n8n's environment or the workflow settings:

```
BLUEPRINT_CHAT_API_BASE=https://chat.blueprintautomation.tech
BLUEPRINT_ADMIN_KEY=your-admin-key
PIPELINE_WEBHOOK_URL=https://your-pipeline-webhook
CRM_WEBHOOK_URL=https://your-crm-webhook
ADMIN_DASHBOARD_URL=https://admin.blueprintautomation.tech
```

## Activation Order

1. `chat-lead-notification` — activate first (needed for testing lead capture)
2. `chat-booking-confirmation` — activate second
3. `chat-tenant-provisioning` — activate when ready to accept APS purchases
4. `chat-daily-digest` — activate last (runs at 8am daily)

## Webhook URLs

After importing and activating, copy each workflow's webhook URL from n8n and set it in Blueprint Chat:

| Workflow | Where to set the URL |
|----------|---------------------|
| Lead Notification | `LEAD_WEBHOOK_URL` in tenant config or per-tenant `lead_webhook_url` |
| Booking Confirmation | `BOOKING_WEBHOOK_URL` in tenant config |
| Provisioning | Configure in your APS purchase flow |
