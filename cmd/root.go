package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/PatrikValkovic/scrappy/internal/cliflags"
	"github.com/PatrikValkovic/scrappy/internal/config"
	"github.com/PatrikValkovic/scrappy/internal/logger"
)

var RootCmd = &cobra.Command{
	Use:           "scrappy",
	Short:         "Scrappy is a tool for scraping web pages",
	SilenceUsage:  true,
	SilenceErrors: true,

	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetEnvPrefix("SCRAPPY")
		viper.AddConfigPath(".")
		viper.SetConfigName("env")
		err := viper.ReadInConfig()
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file %v\n", err)
			return err
		}
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, _ []string) error {
		logger := logger.CreateLogger()
		args, err := config.New()
		if err != nil {
			logger.Infof("Error: %v", err)
			return err
		}
		return startMainLoop(&args, logger)
	},
}

func init() {
	RootCmd.PersistentFlags().Int(cliflags.MaxDepth, 20, "Maximum depth of the crawling")
	RootCmd.PersistentFlags().String(cliflags.ParseRoot, "", "Where to start parsing")
	RootCmd.PersistentFlags().String(cliflags.OutputDir, "./scrapes", "Where to store downloaded files")
	RootCmd.PersistentFlags().String(cliflags.RequiredPrefix, "", "Prefix that all the links must have")
	RootCmd.PersistentFlags().String(cliflags.Environment, "production", "Prefix that all the links must have")
	RootCmd.PersistentFlags().Uint32(cliflags.DownloadConcurrency, 2, "Maximum number of files to download in parallel")
	RootCmd.PersistentFlags().Uint32(cliflags.ParseConcurrency, 2, "Maximum number of files to parse in parallel")
}
