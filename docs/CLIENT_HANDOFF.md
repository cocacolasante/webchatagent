# Blueprint Chat — Client Handoff Guide

Welcome to Blueprint Chat! Your AI assistant is live and ready to help your visitors 24/7.

---

## What Your Chat Bot Does

Your Blueprint Chat AI assistant:

- **Answers questions** about your business instantly, 24/7
- **Captures leads** — when visitors show interest, it collects their name, email, and phone
- **Books appointments** — shows your calendar availability and confirms bookings directly in the chat
- **Represents your brand** — uses your business name, colors, and tone

---

## Adding the Chat Widget to Your Website

You received an embed code that looks like this:

```html
<script src="https://chat.blueprintautomation.tech/widget.js"
        data-tenant-id="YOUR_ID"
        data-position="bottom-right"
        async></script>
```

### WordPress
1. Install the free plugin **"Insert Headers and Footers"**
2. Go to Settings → Insert Headers and Footers
3. Paste your embed code in the **Footer** section
4. Save

### Squarespace
1. Go to **Settings → Advanced → Code Injection**
2. Paste your code in the **Footer** box
3. Save

### Shopify
1. Go to **Online Store → Themes → Edit Code**
2. Open `theme.liquid`
3. Paste before `</body>`
4. Save

### Webflow
1. Go to **Project Settings → Custom Code**
2. Paste in **Footer Code**
3. Publish your site

### Raw HTML
Paste before the `</body>` tag on any page.

---

## Managing Your Bot

Log into your admin dashboard to:
- Update your bot's knowledge base
- View captured leads
- See conversation history
- Change colors and greeting

**Dashboard:** https://admin.blueprintautomation.tech
**Your Tenant ID:** (provided separately)

---

## Updating Your Knowledge Base

If your bot gives an incorrect answer:

1. Log into the admin dashboard
2. Go to **Tenants → [Your Business]**
3. Click the **Knowledge Base** tab
4. Update the relevant FAQ or add a new one
5. Click **Save Changes**

Changes take effect immediately on new conversations.

---

## Viewing Captured Leads

1. Log into the admin dashboard
2. Go to **Tenants → [Your Business]**
3. Click the **Leads** tab

You'll also receive:
- Email notification for each new lead
- Discord/Slack notification (if configured)

---

## Viewing Booked Appointments

Appointments are confirmed directly with your calendar provider (Cal.com, Calendly, or Google Calendar). You'll receive:
- Calendar invitation automatically
- Discord/Slack notification
- Email confirmation to you and the attendee

---

## If Something Isn't Working

1. **Bot gives wrong answers** → Update the Knowledge Base (see above)
2. **Widget not appearing** → Check that the embed code is installed correctly
3. **Leads not arriving** → Check your email spam folder; contact us to verify webhook

**Support:** hello@blueprintautomation.tech
**Response time:** Within 24 hours
