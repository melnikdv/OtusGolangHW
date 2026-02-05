package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/sirupsen/logrus"
)

type handler struct {
	store  storage.Storage
	logger *logrus.Logger
}

func (h *handler) respond(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.WithError(err).Warn("failed to encode response")
		}
	}
}

func (h *handler) error(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		h.respond(w, http.StatusNotFound, map[string]string{"error": "event not found"})
	case errors.Is(err, storage.ErrDateBusy):
		h.respond(w, http.StatusConflict, map[string]string{"error": "time slot is busy"})
	default:
		h.logger.WithError(err).Error("handler error")
		h.respond(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// POST /events
func (h *handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string    `json:"id"`
		Title        string    `json:"title"`
		DateTime     time.Time `json:"datetime"`
		Duration     int64     `json:"duration"` // seconds
		Description  string    `json:"description,omitempty"`
		UserID       string    `json:"user_id"`
		NotifyBefore int64     `json:"notify_before,omitempty"` // seconds
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	event := storage.Event{
		ID:           req.ID,
		Title:        req.Title,
		DateTime:     req.DateTime,
		Duration:     req.Duration,
		Description:  req.Description,
		UserID:       req.UserID,
		NotifyBefore: req.NotifyBefore,
	}

	if err := h.store.Add(event); err != nil {
		h.error(w, err)
		return
	}

	h.respond(w, http.StatusCreated, map[string]string{"id": event.ID})
}

// PUT /events/{id}
func (h *handler) updateEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}

	var req struct {
		Title        string    `json:"title"`
		DateTime     time.Time `json:"datetime"`
		Duration     int64     `json:"duration"`
		Description  string    `json:"description,omitempty"`
		UserID       string    `json:"user_id"`
		NotifyBefore int64     `json:"notify_before,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	event := storage.Event{
		Title:        req.Title,
		DateTime:     req.DateTime,
		Duration:     req.Duration,
		Description:  req.Description,
		UserID:       req.UserID,
		NotifyBefore: req.NotifyBefore,
	}

	if err := h.store.Update(id, event); err != nil {
		h.error(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /events/{id}
func (h *handler) deleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete(id); err != nil {
		h.error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /events/day?date=YYYY-MM-DD
func (h *handler) listDay(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "missing 'date' query param", http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	events, err := h.store.ListDay(date)
	if err != nil {
		h.error(w, err)
		return
	}
	h.respond(w, http.StatusOK, events)
}

// GET /events/week?start=YYYY-MM-DD
func (h *handler) listWeek(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "missing 'start' query param", http.StatusBadRequest)
		return
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	events, err := h.store.ListWeek(start)
	if err != nil {
		h.error(w, err)
		return
	}
	h.respond(w, http.StatusOK, events)
}

// GET /events/month?start=YYYY-MM-DD
func (h *handler) listMonth(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "missing 'start' query param", http.StatusBadRequest)
		return
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	events, err := h.store.ListMonth(start)
	if err != nil {
		h.error(w, err)
		return
	}
	h.respond(w, http.StatusOK, events)
}
