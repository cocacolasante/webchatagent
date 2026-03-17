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

// GoogleCalendarAdapter implements SchedulerAdapter for Google Calendar.
type GoogleCalendarAdapter struct {
	accessToken string
	calendarID  string
	client      *http.Client
}

const googleCalendarAPIBase = "https://www.googleapis.com/calendar/v3"

// NewGoogleCalendarAdapter creates a new Google Calendar adapter.
func NewGoogleCalendarAdapter(accessToken, calendarID string) *GoogleCalendarAdapter {
	return &GoogleCalendarAdapter{
		accessToken: accessToken,
		calendarID:  calendarID,
		client:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *GoogleCalendarAdapter) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// GetAvailability fetches free/busy info from Google Calendar.
func (a *GoogleCalendarAdapter) GetAvailability(ctx context.Context, opts AvailabilityOptions) ([]TimeSlot, error) {
	endpoint := fmt.Sprintf("%s/freeBusy", googleCalendarAPIBase)

	payload := map[string]interface{}{
		"timeMin": opts.StartDate.Format(time.RFC3339),
		"timeMax": opts.EndDate.Format(time.RFC3339),
		"items":   []map[string]string{{"id": a.calendarID}},
	}

	body, _ := json.Marshal(payload)
	req, err := a.newRequest(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google calendar request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google calendar error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Calendars map[string]struct {
			Busy []struct {
				Start string `json:"start"`
				End   string `json:"end"`
			} `json:"busy"`
		} `json:"calendars"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse freeBusy response: %w", err)
	}

	calData, ok := result.Calendars[a.calendarID]
	if !ok {
		return nil, fmt.Errorf("calendar %s not found in response", a.calendarID)
	}

	var slots []TimeSlot
	start := opts.StartDate
	// Generate 30-min slots within business hours (9am-5pm)
	for start.Before(opts.EndDate) {
		hour := start.Hour()
		if hour < 9 || hour >= 17 {
			start = start.Add(30 * time.Minute)
			continue
		}

		end := start.Add(30 * time.Minute)
		available := true
		for _, busy := range calData.Busy {
			busyStart, _ := time.Parse(time.RFC3339, busy.Start)
			busyEnd, _ := time.Parse(time.RFC3339, busy.End)
			if start.Before(busyEnd) && end.After(busyStart) {
				available = false
				break
			}
		}
		if available {
			slots = append(slots, TimeSlot{StartTime: start, EndTime: end, Available: true})
		}
		start = end
	}

	return slots, nil
}

// CreateBooking creates an event on Google Calendar.
func (a *GoogleCalendarAdapter) CreateBooking(ctx context.Context, opts BookingOptions) (*Appointment, error) {
	endpoint := fmt.Sprintf("%s/calendars/%s/events", googleCalendarAPIBase, a.calendarID)

	endTime := opts.SlotStartTime.Add(30 * time.Minute)

	payload := map[string]interface{}{
		"summary":     fmt.Sprintf("Meeting with %s", opts.AttendeeName),
		"description": opts.Notes,
		"start": map[string]string{
			"dateTime": opts.SlotStartTime.Format(time.RFC3339),
			"timeZone": opts.Timezone,
		},
		"end": map[string]string{
			"dateTime": endTime.Format(time.RFC3339),
			"timeZone": opts.Timezone,
		},
		"attendees": []map[string]string{
			{"email": opts.AttendeeEmail, "displayName": opts.AttendeeName},
		},
		"conferenceData": map[string]interface{}{
			"createRequest": map[string]string{
				"requestId": fmt.Sprintf("bp-%d", time.Now().UnixNano()),
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := a.newRequest(ctx, http.MethodPost, endpoint+"?conferenceDataVersion=1", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google calendar create event: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("google calendar error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID             string `json:"id"`
		HtmlLink       string `json:"htmlLink"`
		Start          struct{ DateTime string `json:"dateTime"` } `json:"start"`
		End            struct{ DateTime string `json:"dateTime"` } `json:"end"`
		ConferenceData struct {
			EntryPoints []struct {
				EntryPointType string `json:"entryPointType"`
				URI            string `json:"uri"`
			} `json:"entryPoints"`
		} `json:"conferenceData"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse event response: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, result.Start.DateTime)
	endTimeParsed, _ := time.Parse(time.RFC3339, result.End.DateTime)

	meetingURL := ""
	for _, ep := range result.ConferenceData.EntryPoints {
		if ep.EntryPointType == "video" {
			meetingURL = ep.URI
			break
		}
	}

	return &Appointment{
		ID:         result.ID,
		StartTime:  startTime,
		EndTime:    endTimeParsed,
		ConfirmURL: result.HtmlLink,
		MeetingURL: meetingURL,
	}, nil
}

// CancelBooking deletes a Google Calendar event.
func (a *GoogleCalendarAdapter) CancelBooking(ctx context.Context, appointmentID string) error {
	endpoint := fmt.Sprintf("%s/calendars/%s/events/%s", googleCalendarAPIBase, a.calendarID, appointmentID)
	req, err := a.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("google calendar delete event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("google calendar cancel error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetBooking retrieves a Google Calendar event by ID.
func (a *GoogleCalendarAdapter) GetBooking(ctx context.Context, appointmentID string) (*Appointment, error) {
	endpoint := fmt.Sprintf("%s/calendars/%s/events/%s", googleCalendarAPIBase, a.calendarID, appointmentID)
	req, err := a.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google calendar get event: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google calendar error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID      string `json:"id"`
		HtmlLink string `json:"htmlLink"`
		Start   struct{ DateTime string `json:"dateTime"` } `json:"start"`
		End     struct{ DateTime string `json:"dateTime"` } `json:"end"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse event response: %w", err)
	}

	startTime, _ := time.Parse(time.RFC3339, result.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, result.End.DateTime)

	return &Appointment{
		ID:         result.ID,
		StartTime:  startTime,
		EndTime:    endTime,
		ConfirmURL: result.HtmlLink,
	}, nil
}
