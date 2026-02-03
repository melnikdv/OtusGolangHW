package app

import (
	"context"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
)

type App struct {
	logger Logger
	store  storage.Storage
}

// Совместимость с *logrus.Logger
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

func New(logger Logger, store storage.Storage) *App {
	return &App{logger: logger, store: store}
}

// Методы бизнес-логики — прокси к хранилищу
func (a *App) CreateEvent(_ context.Context, event storage.Event) error {
	return a.store.Add(event)
}

func (a *App) UpdateEvent(_ context.Context, id string, event storage.Event) error {
	return a.store.Update(id, event)
}

func (a *App) DeleteEvent(_ context.Context, id string) error {
	return a.store.Delete(id)
}

func (a *App) ListEventsForDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	return a.store.ListDay(date)
}

func (a *App) ListEventsForWeek(_ context.Context, startDate time.Time) ([]storage.Event, error) {
	return a.store.ListWeek(startDate)
}

func (a *App) ListEventsForMonth(_ context.Context, startDate time.Time) ([]storage.Event, error) {
	return a.store.ListMonth(startDate)
}
