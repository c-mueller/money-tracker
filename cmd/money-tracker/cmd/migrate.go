package cmd

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/internal/repository"

	"github.com/spf13/cobra"
)

var autoApprove bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !autoApprove {
			fmt.Print("Run database migrations? [y/N] ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				logger.Info("migration cancelled")
				return nil
			}
		}

		client, err := repository.NewClient(cfg.Database)
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer client.Close()

		logger.Info("running migrations")
		if err := client.Schema.Create(context.Background()); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}
		logger.Info("migrations completed")
		return nil
	},
}

func init() {
	migrateCmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "skip migration confirmation")
	rootCmd.AddCommand(migrateCmd)
}
