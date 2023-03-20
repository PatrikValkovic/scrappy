package config

import (
	"errors"

	"github.com/spf13/viper"

	"github.com/PatrikValkovic/scrappy/internal/cliflags"
	"github.com/PatrikValkovic/scrappy/internal/environment"
)

type Config struct {
	ParseRoot      string
	OutputDir      string
	MaxDepth       uint64
	RequiredPrefix string
	Environment    string
}

func New() (Config, error) {
	parseRoot := viper.GetString(cliflags.ParseRoot)
	outputDir := viper.GetString(cliflags.OutputDir)
	maxDepth := viper.GetUint64(cliflags.MaxDepth)
	requiredPrefix := viper.GetString(cliflags.RequiredPrefix)

	if parseRoot == "" {
		return Config{}, errors.New("Missing parse root")
	}
	if outputDir == "" {
		return Config{}, errors.New("Missing output dir")
	}
	if err := environment.ValidateEnvironment(viper.GetString(cliflags.Environment)); err != nil {
		return Config{}, errors.New("Invalid environment")
	}
	if requiredPrefix == "" {
		requiredPrefix = parseRoot
	}

	return Config{
		ParseRoot:      parseRoot,
		OutputDir:      outputDir,
		MaxDepth:       maxDepth,
		RequiredPrefix: requiredPrefix,
		Environment:    viper.GetString(cliflags.Environment),
	}, nil
}
