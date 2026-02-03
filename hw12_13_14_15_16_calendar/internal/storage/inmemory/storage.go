package inmemory

import (
	"sync"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) Add(event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.events {
		if e.UserID == event.UserID && e.DateTime.Equal(event.DateTime) {
			return storage.ErrDateBusy
		}
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) Update(id string, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}
	s.events[id] = event
	return nil
}

func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}
	delete(s.events, id)
	return nil
}

func (s *Storage) ListDay(date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	for _, e := range s.events {
		if e.DateTime.After(start) && e.DateTime.Before(end) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Storage) ListWeek(startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := start.Add(7 * 24 * time.Hour)

	for _, e := range s.events {
		if e.DateTime.After(start) && e.DateTime.Before(end) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Storage) ListMonth(startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	start := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	end := start.AddDate(0, 1, 0)

	for _, e := range s.events {
		if e.DateTime.After(start) && e.DateTime.Before(end) {
			result = append(result, e)
		}
	}
	return result, nil
}
