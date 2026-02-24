package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateEvent(t *testing.T) {
	store := inmemory.New()
	logg := logrus.New()
	h := &handler{store: store, logger: logg}

	body := `{
		"id": "1",
		"title": "Test Event",
		"datetime": "2026-02-10T10:00:00Z",
		"duration": 3600,
		"user_id": "user1"
	}`

	req := httptest.NewRequest("POST", "/events", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.createEvent(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp["id"])
}

func TestListEventsForDay(t *testing.T) {
	store := inmemory.New()
	logg := logrus.New()
	h := &handler{store: store, logger: logg}

	// Создаём событие
	event := storage.Event{
		ID:       "1",
		Title:    "Test",
		DateTime: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
		Duration: 3600,
		UserID:   "user1",
	}
	assert.NoError(t, store.Add(event))

	req := httptest.NewRequest("GET", "/events/day?date=2026-02-10", nil)
	w := httptest.NewRecorder()

	h.listDay(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var events []storage.Event
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &events))
	assert.Len(t, events, 1)
	assert.Equal(t, "Test", events[0].Title)
}
