# Plugin Interface — Blueprint Chat

Plugins extend Blueprint Chat without modifying core code. Each plugin is a Go struct that implements the `Plugin` interface.

## The Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string

    OnMessageReceived(ctx context.Context, msg *chat.Message, t *tenant.Tenant) error
    OnResponseGenerated(ctx context.Context, resp *ChatResponse, t *tenant.Tenant) error
    OnLeadCaptured(ctx context.Context, lead *leads.Lead, t *tenant.Tenant) error
    OnAppointmentBooked(ctx context.Context, appt *scheduler.Appointment, t *tenant.Tenant) error
    OnSessionStarted(ctx context.Context, session *chat.Session, t *tenant.Tenant) error
    OnSessionEnded(ctx context.Context, session *chat.Session, t *tenant.Tenant) error
}
```

## Hook Reference

| Hook | When it fires | Common use |
|------|--------------|------------|
| `OnMessageReceived` | Every user message, before Claude processes it | Content filtering, logging, custom routing |
| `OnResponseGenerated` | After Claude responds, before sending to widget | Response modification, analytics |
| `OnLeadCaptured` | When lead form is submitted | CRM sync, Slack notification |
| `OnAppointmentBooked` | When booking is confirmed | Calendar sync, CRM update |
| `OnSessionStarted` | First message in a new session | UTM tracking, A/B testing |
| `OnSessionEnded` | Session TTL expires | Conversation archiving |

## Creating a Plugin

### Step 1: Create your plugin file

```go
// internal/plugins/mycompany/mycompany.go
package mycompany

import (
    "context"
    "github.com/blueprintautomation/blueprint-chat/internal/chat"
    "github.com/blueprintautomation/blueprint-chat/internal/leads"
    "github.com/blueprintautomation/blueprint-chat/internal/plugins"
    "github.com/blueprintautomation/blueprint-chat/internal/scheduler"
    "github.com/blueprintautomation/blueprint-chat/internal/tenant"
)

type MyPlugin struct {
    webhookURL string
}

func New(webhookURL string) *MyPlugin {
    return &MyPlugin{webhookURL: webhookURL}
}

// Compile-time check
var _ plugins.Plugin = (*MyPlugin)(nil)

func (p *MyPlugin) Name() string    { return "mycompany" }
func (p *MyPlugin) Version() string { return "1.0.0" }

func (p *MyPlugin) OnLeadCaptured(ctx context.Context, lead *leads.Lead, t *tenant.Tenant) error {
    // Your custom logic here
    return nil
}

// Implement remaining hooks with nil returns to skip them
func (p *MyPlugin) OnMessageReceived(_ context.Context, _ *chat.Message, _ *tenant.Tenant) error { return nil }
func (p *MyPlugin) OnResponseGenerated(_ context.Context, _ *plugins.ChatResponse, _ *tenant.Tenant) error { return nil }
func (p *MyPlugin) OnAppointmentBooked(_ context.Context, _ *scheduler.Appointment, _ *tenant.Tenant) error { return nil }
func (p *MyPlugin) OnSessionStarted(_ context.Context, _ *chat.Session, _ *tenant.Tenant) error { return nil }
func (p *MyPlugin) OnSessionEnded(_ context.Context, _ *chat.Session, _ *tenant.Tenant) error { return nil }
```

### Step 2: Register in main.go

```go
// In cmd/server/main.go, after creating registry:
pluginRegistry := plugins.NewRegistry(log)
pluginRegistry.Register(mycompany.New(cfg.MyWebhookURL))
```

### Step 3: Fire hooks (already wired in)

Hooks are fired from handlers automatically — you don't need to modify handler code.

## Example: LoggingPlugin

Logs all messages to a file:

```go
type LoggingPlugin struct {
    file *os.File
    mu   sync.Mutex
}

func (p *LoggingPlugin) OnMessageReceived(ctx context.Context, msg *chat.Message, t *tenant.Tenant) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    fmt.Fprintf(p.file, "[%s] tenant=%s role=%s len=%d\n",
        time.Now().Format(time.RFC3339), t.ID, msg.Role, len(msg.Content))
    return nil
}
```

## Future Plugin Stubs

These plugins are planned but not yet built:

- **VoicePlugin** — `OnMessageReceived`: convert speech to text input
- **LiveHandoffPlugin** — `OnMessageReceived`: detect escalation phrases, open live chat channel
- **PaymentPlugin** — `OnResponseGenerated`: inject payment link when pricing discussed
- **ProductLookupPlugin** — `OnMessageReceived`: query product catalog, inject results into context
- **TranslationPlugin** — `OnMessageReceived`/`OnResponseGenerated`: detect language, translate
- **AnalyticsPlugin** — all hooks: send events to Mixpanel/PostHog/GA4
