package scheduler

import (
	"context"
	"time"
)

// SchedulerAdapter is the interface all scheduling providers implement.
type SchedulerAdapter interface {
	GetAvailability(ctx context.Context, opts AvailabilityOptions) ([]TimeSlot, error)
	CreateBooking(ctx context.Context, opts BookingOptions) (*Appointment, error)
	CancelBooking(ctx context.Context, appointmentID string) error
	GetBooking(ctx context.Context, appointmentID string) (*Appointment, error)
}

// AvailabilityOptions specifies parameters for fetching available time slots.
type AvailabilityOptions struct {
	EventTypeID string
	StartDate   time.Time
	EndDate     time.Time
	Timezone    string
}

// BookingOptions specifies parameters for creating a booking.
type BookingOptions struct {
	SlotStartTime time.Time `json:"slotStartTime"`
	AttendeeEmail string    `json:"attendeeEmail"`
	AttendeeName  string    `json:"attendeeName"`
	Notes         string    `json:"notes"`
	EventTypeID   string    `json:"eventTypeId"`
	Timezone      string    `json:"timezone"`
}

// TimeSlot represents a single available (or busy) time window.
type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Available bool      `json:"available"`
}

// Appointment is the result of a successful booking.
type Appointment struct {
	ID         string    `json:"id"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	ConfirmURL string    `json:"confirmUrl"`
	CancelURL  string    `json:"cancelUrl"`
	MeetingURL string    `json:"meetingUrl"`
}
