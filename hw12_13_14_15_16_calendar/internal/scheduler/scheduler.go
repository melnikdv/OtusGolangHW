package scheduler

import (
	"context"
	"time"

	_ "github.com/lib/pq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/rmq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	store     storage.Storage
	publisher rmq.Publisher
	interval  time.Duration
	queue     string
	logger    *logrus.Logger
}

func New(
	logger *logrus.Logger,
	store storage.Storage,
	publisher rmq.Publisher,
	interval time.Duration,
	queue string,
) *Scheduler {
	return &Scheduler{
		logger:    logger,
		store:     store,
		publisher: publisher,
		interval:  interval,
		queue:     queue,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processEvents(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Scheduler) processEvents(_ context.Context) {
	events, err := s.store.ListUpcomingReminders(time.Now())
	if err != nil {
		return
	}

	for _, ev := range events {
		msg := rmq.Message{
			EventID: ev.ID,
			Title:   ev.Title,
			UserID:  ev.UserID,
			At:      ev.DateTime.Format(time.RFC3339),
		}
		if err := s.publisher.Publish(s.queue, msg); err == nil {
			// ✅ Помечаем как уведомлённое
			ev.Notified = true
			if err := s.store.Update(ev.ID, ev); err != nil {
				s.logger.WithError(err).Warn("failed to mark event as notified")
			}
		}
	}

	// Очистка
	if err := s.store.CleanupOldEvents(time.Now().AddDate(-1, 0, 0)); err != nil {
		s.logger.WithError(err).Warn("failed to cleanup old events")
	}
}
