package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/PatrikValkovic/scrappy/internal/args"
	"github.com/PatrikValkovic/scrappy/internal/cliflags"
	"github.com/PatrikValkovic/scrappy/internal/logger"
)

var RootCmd = &cobra.Command{
	Use:   "scrappy",
	Short: "Scrappy is a tool for scraping web pages",

	RunE: func(cmd *cobra.Command, _ []string) error {
		logger := logger.CreateLogger()
		args := args.GetArgs(cmd)
		startMainLoop(&args, logger)
		return nil
	},

	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.AddConfigPath(".")
		viper.SetConfigName("env")
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		viper.AutomaticEnv()
		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().Int(cliflags.MaxDepth, 20, "Maximum depth of the crawling")
	RootCmd.PersistentFlags().String(cliflags.ParseRoot, "", "Where to start parsing")
	RootCmd.PersistentFlags().String(cliflags.OutputDir, "", "Where to store downloaded files")
	RootCmd.PersistentFlags().String(cliflags.RequiredPrefix, "", "Prefix that all the links must have")
	RootCmd.PersistentFlags().String(cliflags.Environment, "production", "Prefix that all the links must have")
}
