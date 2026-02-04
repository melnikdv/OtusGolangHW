package storage

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("date is already occupied")
)

type Event struct {
	ID           string    `db:"id"`
	Title        string    `db:"title"`
	DateTime     time.Time `db:"datetime"`
	Duration     int64     `db:"duration"` // seconds
	Description  string    `db:"description"`
	UserID       string    `db:"user_id"`
	NotifyBefore int64     `db:"notify_before"` // seconds
	Notified     bool      `db:"notified"`
}

type Storage interface {
	Add(event Event) error
	Update(id string, event Event) error
	Delete(id string) error
	ListDay(date time.Time) ([]Event, error)
	ListWeek(startDate time.Time) ([]Event, error)
	ListMonth(startDate time.Time) ([]Event, error)
	ListUpcomingReminders(now time.Time) ([]Event, error)
	CleanupOldEvents(before time.Time) error
}
