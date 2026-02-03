package sqlstorage

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/migration"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	db *sqlx.DB
}

func New(dsn string, logger *logrus.Logger) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	logger.Info("Applying database migrations")
	applied, err := migration.Apply(db)
	if err != nil {
		logger.WithError(err).Errorf("Migration '%s' failed", migration.MigrationName)
		return nil, err
	}

	if applied {
		logger.Infof("Migration '%s' - applied successfully", migration.MigrationName)
	} else {
		logger.Infof("Migration '%s' - already applied", migration.MigrationName)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Add(event storage.Event) error {
	query := `
		INSERT INTO events (id, title, datetime, duration, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.Exec(query,
		event.ID,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.NotifyBefore,
	)
	return err
}

func (s *Storage) Update(id string, event storage.Event) error {
	query := `
		UPDATE events
		SET title = $1, datetime = $2, duration = $3, description = $4, user_id = $5, notify_before = $6
		WHERE id = $7`
	_, err := s.db.Exec(query,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.NotifyBefore,
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
		SELECT id, title, datetime, duration, description, user_id, notify_before
		FROM events
		WHERE datetime > $1 AND datetime < $2`
	err := s.db.Select(&events, query, start, end)
	return events, err
}
