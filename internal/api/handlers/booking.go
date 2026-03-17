package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	mw "github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/scheduler"
	"go.uber.org/zap"
)

// BookingHandler handles appointment booking requests.
type BookingHandler struct {
	calcomBase   string
	calendlyBase string
	encKey       string
	log          *zap.Logger
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(calcomBase, calendlyBase, encKey string, log *zap.Logger) *BookingHandler {
	return &BookingHandler{
		calcomBase:   calcomBase,
		calendlyBase: calendlyBase,
		encKey:       encKey,
		log:          log,
	}
}

func (h *BookingHandler) getAdapter(t interface{ GetSchedulerType() string; GetSchedulerAPIKey() string; GetSchedulerCalendarID() string }) (scheduler.SchedulerAdapter, error) {
	return nil, nil // placeholder — actual call below
}

// GetAvailability handles POST /api/booking/availability.
func (h *BookingHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	t := mw.TenantFromContext(r.Context())
	if t == nil || t.SchedulerType == "" {
		http.Error(w, `{"error":"scheduler not configured"}`, http.StatusBadRequest)
		return
	}

	apiBase := h.calcomBase
	if t.SchedulerType == "calendly" {
		apiBase = h.calendlyBase
	}

	adapter, err := scheduler.NewAdapter(t.SchedulerType, t.SchedulerConfig.APIKey, apiBase, t.SchedulerConfig.CalendarID)
	if err != nil {
		h.log.Error("create scheduler adapter", zap.Error(err))
		http.Error(w, `{"error":"scheduler configuration error"}`, http.StatusInternalServerError)
		return
	}

	startDate := time.Now().UTC()
	endDate := startDate.Add(7 * 24 * time.Hour)

	slots, err := adapter.GetAvailability(r.Context(), scheduler.AvailabilityOptions{
		EventTypeID: t.SchedulerConfig.EventTypeID,
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    r.URL.Query().Get("timezone"),
	})
	if err != nil {
		h.log.Error("get availability", zap.Error(err))
		http.Error(w, `{"error":"failed to fetch availability"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"slots": slots})
}

// CreateBooking handles POST /api/booking/create.
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	t := mw.TenantFromContext(r.Context())
	if t == nil || t.SchedulerType == "" {
		http.Error(w, `{"error":"scheduler not configured"}`, http.StatusBadRequest)
		return
	}

	var opts scheduler.BookingOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	opts.EventTypeID = t.SchedulerConfig.EventTypeID

	apiBase := h.calcomBase
	if t.SchedulerType == "calendly" {
		apiBase = h.calendlyBase
	}

	adapter, err := scheduler.NewAdapter(t.SchedulerType, t.SchedulerConfig.APIKey, apiBase, t.SchedulerConfig.CalendarID)
	if err != nil {
		h.log.Error("create scheduler adapter", zap.Error(err))
		http.Error(w, `{"error":"scheduler configuration error"}`, http.StatusInternalServerError)
		return
	}

	appt, err := adapter.CreateBooking(r.Context(), opts)
	if err != nil {
		h.log.Error("create booking", zap.Error(err))
		http.Error(w, `{"error":"failed to create booking"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(appt)
}
