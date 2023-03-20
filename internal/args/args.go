package args

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/PatrikValkovic/scrappy/internal/cliflags"
)

type Args struct {
	ParseRoot      string
	OutputDir      string
	MaxDepth       uint64
	RequiredPrefix string
}

func GetArgs(cmd *cobra.Command) Args {
	parseRoot := viper.GetString(cliflags.ParseRoot)
	outputDir := viper.GetString(cliflags.OutputDir)
	maxDepth := viper.GetUint64(cliflags.MaxDepth)
	requiredPrefix := viper.GetString(cliflags.RequiredPrefix)

	if parseRoot == "" {
		fmt.Println("Missing parse root")
		cmd.Help()
		os.Exit(1)
	}
	if outputDir == "" {
		fmt.Println("Missing output dir")
		cmd.Help()
		os.Exit(1)
	}
	if requiredPrefix == "" {
		requiredPrefix = parseRoot
	}

	return Args{
		ParseRoot:      parseRoot,
		OutputDir:      outputDir,
		MaxDepth:       maxDepth,
		RequiredPrefix: requiredPrefix,
	}
}
