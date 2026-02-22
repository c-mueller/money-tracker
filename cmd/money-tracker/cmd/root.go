package cmd

import (
	"icekalt.dev/money-tracker/internal/config"
	"icekalt.dev/money-tracker/internal/logging"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	cfgFile string
	cfg     config.Config
	logger  *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "money-tracker",
	Short: "Deterministic household budget tracker",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}

		logger, err = logging.New(cfg.Logging.Level)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
}

func Execute() error {
	return rootCmd.Execute()
}
