package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/config"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/rmq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/spf13/cobra"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "calendar_scheduler",
		Short: "Calendar scheduler for sending reminders via RabbitMQ",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.LoadSchedulerConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			logg := logger.New(cfg.Logger.Level)

			store, err := sqlstorage.New(cfg.Storage.SQL.DSN, logg)
			if err != nil {
				return fmt.Errorf("failed to init SQL storage: %w", err)
			}

			publisher, err := rmq.NewAMQP(cfg.RMQ.URL)
			if err != nil {
				return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
			}
			defer func() { _ = publisher.Close() }()

			interval, err := time.ParseDuration(cfg.Interval)
			if err != nil {
				return fmt.Errorf("invalid interval: %w", err)
			}

			sched := scheduler.New(logg, store, publisher, interval, cfg.RMQ.Queue)

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer func() {
				cancel()
				logg.Info("shutting down calendar_scheduler")
			}()

			logg.Info("Starting calendar scheduler...")
			return sched.Run(ctx)
		},
		SilenceUsage: true,
	}

	rootCmd.Flags().StringVar(&configPath, "config", "configs/scheduler.yaml", "path to config file")
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
