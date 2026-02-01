package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/config"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/app"
	"github.com/spf13/cobra"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar service skeleton",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			app := app.New(cfg)
			return app.Run()
		},
	}

	// Подкоманда `version`
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(_ *cobra.Command, _ []string) {
			printVersion()
		},
	}

	rootCmd.Flags().StringVar(&configPath, "config", "config.yaml", "path to config file")
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
