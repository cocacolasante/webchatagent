package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CalendlyAdapter implements SchedulerAdapter for Calendly.
type CalendlyAdapter struct {
	apiKey  string
	apiBase string
	client  *http.Client
}

// NewCalendlyAdapter creates a new Calendly scheduler adapter.
func NewCalendlyAdapter(apiKey, apiBase string) *CalendlyAdapter {
	return &CalendlyAdapter{
		apiKey:  apiKey,
		apiBase: apiBase,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *CalendlyAdapter) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// GetAvailability fetches available time slots from Calendly.
func (a *CalendlyAdapter) GetAvailability(ctx context.Context, opts AvailabilityOptions) ([]TimeSlot, error) {
	endpoint := fmt.Sprintf("%s/event_type_available_times", a.apiBase)

	req, err := a.newRequest(ctx, http.MethodGet, fmt.Sprintf(
		"%s?event_type=%s&start_time=%s&end_time=%s",
		endpoint,
		opts.EventTypeID,
		opts.StartDate.Format(time.RFC3339),
		opts.EndDate.Format(time.RFC3339),
	), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calendly availability request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calendly error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Collection []struct {
			StartTime string `json:"start_time"`
			Status    string `json:"status"`
		} `json:"collection"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse availability response: %w", err)
	}

	var slots []TimeSlot
	for _, item := range result.Collection {
		startTime, err := time.Parse(time.RFC3339, item.StartTime)
		if err != nil {
			continue
		}
		slots = append(slots, TimeSlot{
			StartTime: startTime,
			EndTime:   startTime.Add(30 * time.Minute),
			Available: item.Status == "available",
		})
	}

	return slots, nil
}

// CreateBooking creates a new booking via Calendly.
func (a *CalendlyAdapter) CreateBooking(ctx context.Context, opts BookingOptions) (*Appointment, error) {
	// Calendly requires invitee scheduling — this is a simplified implementation
	payload := map[string]interface{}{
		"event_type_uuid": opts.EventTypeID,
		"start_time":      opts.SlotStartTime.Format(time.RFC3339),
		"invitee": map[string]interface{}{
			"name":  opts.AttendeeName,
			"email": opts.AttendeeEmail,
		},
	}

	body, _ := json.Marshal(payload)
	req, err := a.newRequest(ctx, http.MethodPost, fmt.Sprintf("%s/one_off_event_types", a.apiBase), strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calendly booking request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("calendly error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Resource struct {
			URI       string `json:"uri"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
			Location  struct {
				JoinURL string `json:"join_url"`
			} `json:"location"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse booking response: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, result.Resource.StartTime)
	endTime, _ := time.Parse(time.RFC3339, result.Resource.EndTime)

	return &Appointment{
		ID:         result.Resource.URI,
		StartTime:  startTime,
		EndTime:    endTime,
		MeetingURL: result.Resource.Location.JoinURL,
	}, nil
}

// CancelBooking cancels a Calendly booking.
func (a *CalendlyAdapter) CancelBooking(ctx context.Context, appointmentID string) error {
	req, err := a.newRequest(ctx, http.MethodPost,
		fmt.Sprintf("%s/invitee_no_shows", a.apiBase),
		strings.NewReader(fmt.Sprintf(`{"invitee":"%s"}`, appointmentID)))
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("calendly cancel request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("calendly cancel error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetBooking retrieves a Calendly booking by URI.
func (a *CalendlyAdapter) GetBooking(ctx context.Context, appointmentID string) (*Appointment, error) {
	req, err := a.newRequest(ctx, http.MethodGet,
		fmt.Sprintf("%s/scheduled_events/%s", a.apiBase, appointmentID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calendly get booking request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calendly error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Resource struct {
			URI       string `json:"uri"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse booking response: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, result.Resource.StartTime)
	endTime, _ := time.Parse(time.RFC3339, result.Resource.EndTime)

	return &Appointment{
		ID:        result.Resource.URI,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}
