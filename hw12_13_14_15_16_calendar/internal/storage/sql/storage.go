package sqlstorage

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	db *sqlx.DB
}

func New(dsn string, _ *logrus.Logger) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Add(event storage.Event) error {
	query := `
		INSERT INTO events (id, title, datetime, duration, description, user_id, notify_before, notified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.Exec(query,
		event.ID,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.NotifyBefore,
		event.Notified,
	)
	if err != nil {
		// Проверяем, является ли ошибка нарушением уникального ограничения
		if pqErr, ok := err.(*pq.Error); ok {
			// Код 23505 = unique_violation
			if pqErr.Code == "23505" {
				return storage.ErrDateBusy
			}
		}
		return err
	}
	return nil
}

func (s *Storage) Update(id string, event storage.Event) error {
	query := `
		UPDATE events
		SET title = $1, datetime = $2, duration = $3, description = $4, user_id = $5, notify_before = $6, notified = $7
		WHERE id = $8`
	_, err := s.db.Exec(query,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.NotifyBefore,
		event.Notified,
		id,
	)
	return err
}

func (s *Storage) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM events WHERE id = $1", id)
	return err
}

func (s *Storage) ListDay(date time.Time) ([]storage.Event, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	return s.listBetween(start, end)
}

func (s *Storage) ListWeek(startDate time.Time) ([]storage.Event, error) {
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := start.Add(7 * 24 * time.Hour)
	return s.listBetween(start, end)
}

func (s *Storage) ListMonth(startDate time.Time) ([]storage.Event, error) {
	start := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	end := start.AddDate(0, 1, 0)
	return s.listBetween(start, end)
}

func (s *Storage) listBetween(start, end time.Time) ([]storage.Event, error) {
	var events []storage.Event
	query := `
		SELECT id, title, datetime, duration, description, user_id, notify_before, notified
		FROM events
		WHERE datetime > $1 AND datetime < $2`
	err := s.db.Select(&events, query, start, end)
	return events, err
}

func (s *Storage) ListUpcomingReminders(now time.Time) ([]storage.Event, error) {
	query := `
		SELECT id, title, datetime, duration, description, user_id, notify_before, notified
		FROM events
		WHERE notify_before > 0
		  AND datetime - make_interval(secs => notify_before) <= $1
		  AND datetime > $1
		  AND notified = false
	`
	rows, err := s.db.Query(query, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []storage.Event
	for rows.Next() {
		var ev storage.Event
		err := rows.Scan(
			&ev.ID,
			&ev.Title,
			&ev.DateTime,
			&ev.Duration,
			&ev.Description,
			&ev.UserID,
			&ev.NotifyBefore,
			&ev.Notified,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, nil
}

func (s *Storage) CleanupOldEvents(before time.Time) error {
	_, err := s.db.Exec("DELETE FROM events WHERE datetime < $1", before)
	return err
}

// DB возвращает подключение к базе данных (для миграций)
func (s *Storage) DB() *sqlx.DB {
	return s.db
}
