package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/config"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/rmq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/sender"
	"github.com/spf13/cobra"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "calendar_sender",
		Short: "Notification sender for Calendar service",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.LoadSenderConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			logg := logger.New(cfg.Logger.Level)

			consumer, err := rmq.NewAMQP(cfg.RMQ.URL)
			if err != nil {
				return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
			}
			defer func() { _ = consumer.Close() }()

			sndr := sender.New(consumer, logg, cfg.RMQ.Queue)

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer func() {
				cancel()
				logg.Info("shutting down calendar_sender")
			}()

			return sndr.Run(ctx)
		},
		SilenceUsage: true,
	}

	rootCmd.Flags().StringVar(&configPath, "config", "configs/sender.yaml", "path to config file")
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
