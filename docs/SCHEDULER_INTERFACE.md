# Scheduler Interface — Blueprint Chat

## The SchedulerAdapter Interface

```go
type SchedulerAdapter interface {
    GetAvailability(ctx context.Context, opts AvailabilityOptions) ([]TimeSlot, error)
    CreateBooking(ctx context.Context, opts BookingOptions) (*Appointment, error)
    CancelBooking(ctx context.Context, appointmentID string) error
    GetBooking(ctx context.Context, appointmentID string) (*Appointment, error)
}
```

## Adding a New Scheduler (e.g., Acuity Scheduling)

1. Create `internal/scheduler/acuity.go`
2. Implement all 4 interface methods
3. Add a case to `factory.go`:
```go
case "acuity":
    return NewAcuityAdapter(apiKey, apiBase), nil
```
4. Add `"acuity"` as an option in the admin dashboard scheduler selector

## Cal.com Setup

1. Log into app.cal.com
2. Go to **Settings → Developer → API Keys**
3. Create a new API key with full access
4. Copy the key
5. Get your Event Type ID from the URL when editing an event type: `app.cal.com/event-types/12345` → ID is `12345`
6. Set in tenant config: `schedulerType = "calcom"`, paste API key and event type ID

## Calendly Setup

1. Log into calendly.com
2. Go to **Integrations → API & Webhooks**
3. Generate a Personal Access Token
4. Copy your Event Type URL (the UUID portion)
5. Set in tenant config: `schedulerType = "calendly"`

## Google Calendar Setup

1. Go to console.cloud.google.com
2. Create a project → Enable Google Calendar API
3. Create OAuth 2.0 credentials → Download JSON
4. Exchange for an access token (requires OAuth flow — contact Blueprint Automation for setup)
5. Set in tenant config: `schedulerType = "google"`, provide access token and calendar ID
