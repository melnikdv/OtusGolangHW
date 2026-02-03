package app

import (
	"fmt"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/config"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/server"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	sqlstorage "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/sirupsen/logrus"
)

type App struct {
	cfg    *config.Config
	logger *logrus.Logger
	store  storage.Storage
	srv    *server.Server
}

func New(cfg *config.Config) *App {
	log := logger.New(cfg.Logger.Level)
	var store storage.Storage

	switch cfg.Storage.Type {
	case config.InMemory:
		store = inmemory.New()
	case config.SQL:
		s, err := sqlstorage.New(cfg.Storage.SQL.DSN, log)
		if err != nil {
			panic(fmt.Sprintf("failed to init SQL storage: %v", err))
		}
		store = s
	default:
		panic("unknown storage type")
	}

	srv := server.New(log, cfg.Server.Host, cfg.Server.Port)
	return &App{
		cfg:    cfg,
		logger: log,
		store:  store,
		srv:    srv,
	}
}

func (a *App) Run() error {
	return a.srv.Start()
}
