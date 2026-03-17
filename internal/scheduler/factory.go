package scheduler

import (
	"fmt"
)

// NewAdapter creates the appropriate SchedulerAdapter based on the scheduler type.
func NewAdapter(schedulerType, apiKey, apiBase, calendarID string) (SchedulerAdapter, error) {
	switch schedulerType {
	case "calcom":
		if apiBase == "" {
			apiBase = "https://api.cal.com/v1"
		}
		return NewCalComAdapter(apiKey, apiBase), nil
	case "calendly":
		if apiBase == "" {
			apiBase = "https://api.calendly.com"
		}
		return NewCalendlyAdapter(apiKey, apiBase), nil
	case "google":
		return NewGoogleCalendarAdapter(apiKey, calendarID), nil
	default:
		return nil, fmt.Errorf("unsupported scheduler type: %s", schedulerType)
	}
}
