package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CalComAdapter implements SchedulerAdapter for Cal.com.
type CalComAdapter struct {
	apiKey  string
	apiBase string
	client  *http.Client
}

// NewCalComAdapter creates a new Cal.com scheduler adapter.
func NewCalComAdapter(apiKey, apiBase string) *CalComAdapter {
	return &CalComAdapter{
		apiKey:  apiKey,
		apiBase: apiBase,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// GetAvailability fetches available time slots from Cal.com.
func (a *CalComAdapter) GetAvailability(ctx context.Context, opts AvailabilityOptions) ([]TimeSlot, error) {
	endpoint := fmt.Sprintf("%s/availability", a.apiBase)

	params := url.Values{}
	params.Set("eventTypeId", opts.EventTypeID)
	params.Set("startTime", opts.StartDate.Format(time.RFC3339))
	params.Set("endTime", opts.EndDate.Format(time.RFC3339))
	if opts.Timezone != "" {
		params.Set("timeZone", opts.Timezone)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cal.com availability request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cal.com error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Busy []struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"busy"`
		TimeZone string `json:"timeZone"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse availability response: %w", err)
	}

	// Convert busy windows into 30-min available slots
	var slots []TimeSlot
	start := opts.StartDate
	for start.Before(opts.EndDate) {
		end := start.Add(30 * time.Minute)
		available := true
		for _, busy := range result.Busy {
			busyStart, _ := time.Parse(time.RFC3339, busy.Start)
			busyEnd, _ := time.Parse(time.RFC3339, busy.End)
			if start.Before(busyEnd) && end.After(busyStart) {
				available = false
				break
			}
		}
		if available {
			slots = append(slots, TimeSlot{
				StartTime: start,
				EndTime:   end,
				Available: true,
			})
		}
		start = end
	}

	return slots, nil
}

// CreateBooking creates a new booking via Cal.com.
func (a *CalComAdapter) CreateBooking(ctx context.Context, opts BookingOptions) (*Appointment, error) {
	endpoint := fmt.Sprintf("%s/bookings", a.apiBase)

	payload := map[string]interface{}{
		"eventTypeId": opts.EventTypeID,
		"start":       opts.SlotStartTime.Format(time.RFC3339),
		"responses": map[string]interface{}{
			"name":  opts.AttendeeName,
			"email": opts.AttendeeEmail,
			"notes": opts.Notes,
		},
		"timeZone": opts.Timezone,
		"language": "en",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cal.com booking request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cal.com error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID           int    `json:"id"`
		UID          string `json:"uid"`
		StartTime    string `json:"startTime"`
		EndTime      string `json:"endTime"`
		MeetingURL   string `json:"videoCallData.url"`
		CancellationReason string `json:"cancellationReason"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse booking response: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, result.StartTime)
	endTime, _ := time.Parse(time.RFC3339, result.EndTime)

	return &Appointment{
		ID:         result.UID,
		StartTime:  startTime,
		EndTime:    endTime,
		ConfirmURL: fmt.Sprintf("https://cal.com/booking/%s", result.UID),
		CancelURL:  fmt.Sprintf("https://cal.com/cancelling/%s", result.UID),
		MeetingURL: result.MeetingURL,
	}, nil
}

// CancelBooking cancels a booking via Cal.com.
func (a *CalComAdapter) CancelBooking(ctx context.Context, appointmentID string) error {
	endpoint := fmt.Sprintf("%s/bookings/%s/cancel", a.apiBase, appointmentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("cal.com cancel request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cal.com cancel error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetBooking retrieves a booking by ID from Cal.com.
func (a *CalComAdapter) GetBooking(ctx context.Context, appointmentID string) (*Appointment, error) {
	endpoint := fmt.Sprintf("%s/bookings/%s", a.apiBase, appointmentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cal.com get booking request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cal.com error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		UID       string `json:"uid"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse booking response: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, result.StartTime)
	endTime, _ := time.Parse(time.RFC3339, result.EndTime)

	return &Appointment{
		ID:        result.UID,
		StartTime: startTime,
		EndTime:   endTime,
		CancelURL: fmt.Sprintf("https://cal.com/cancelling/%s", result.UID),
	}, nil
}
