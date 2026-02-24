package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/config"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/migration"
	grpcserver "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/server/grpc"
	httpserver "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	sqlstorage "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/spf13/cobra"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar service with HTTP and gRPC API",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			logg := logger.New(cfg.Logger.Level)

			var store storage.Storage
			switch cfg.Storage.Type {
			case config.InMemory:
				store = inmemory.New()
			case config.SQL:
				s, err := sqlstorage.New(cfg.Storage.SQL.DSN, logg)
				if err != nil {
					return fmt.Errorf("failed to init SQL storage: %w", err)
				}
				// Применяем миграции только в calendar
				logg.Info("Applying database migrations")
				applied, err := migration.Apply(s.DB())
				if err != nil {
					logg.WithError(err).Errorf("Migration '%s' failed", migration.MigrationName)
					return fmt.Errorf("failed to apply migration: %w", err)
				}
				if applied {
					logg.Infof("Migration '%s' - applied successfully", migration.MigrationName)
				} else {
					logg.Infof("Migration '%s' - already applied", migration.MigrationName)
				}
				store = s
			default:
				return fmt.Errorf("unknown storage type: %s", cfg.Storage.Type)
			}

			// Создаём серверы
			httpSrv := httpserver.New(logg, store, cfg.Server.Host, cfg.Server.Port)
			grpcSrv := grpcserver.New(logg, store, cfg.Server.Host, cfg.Server.GRPCPort)

			ctx, cancel := signal.NotifyContext(context.Background(),
				syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			defer cancel()

			// Горутина для graceful shutdown
			go func() {
				<-ctx.Done()
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer shutdownCancel()

				if err := httpSrv.Stop(shutdownCtx); err != nil {
					logg.WithError(err).Warn("HTTP shutdown error")
				}
				if err := grpcSrv.Stop(shutdownCtx); err != nil {
					logg.WithError(err).Warn("gRPC shutdown error")
				}
			}()

			// Запускаем HTTP в фоне
			go func() {
				if err := httpSrv.Start(ctx); err != nil && err != context.Canceled {
					logg.WithError(err).Error("HTTP server failed")
					cancel()
				}
			}()

			// Блокируемся на gRPC (можно и наоборот)
			if err := grpcSrv.Start(ctx); err != nil && err != context.Canceled {
				logg.WithError(err).Error("gRPC server failed")
				return err
			}

			return nil
		},
		SilenceUsage: true,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(_ *cobra.Command, _ []string) {
			printVersion()
		},
	}

	rootCmd.Flags().StringVar(&configPath, "config", "configs/calendar.yaml", "path to config file")
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
