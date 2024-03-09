package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"regexp"

	"github.com/PatrikValkovic/scrappy/internal/cliflags"
	"github.com/PatrikValkovic/scrappy/internal/environment"
)

type Config struct {
	ParseRoot           string
	OutputDir           string
	MaxDepth            uint64
	RequiredPrefix      string
	Environment         string
	DownloadConcurrency uint32
	ParseConcurrency    uint32
	IgnorePatterns      []*regexp.Regexp
}

func New() (Config, error) {
	parseRoot := viper.GetString(cliflags.ParseRoot)
	outputDir := viper.GetString(cliflags.OutputDir)
	maxDepth := viper.GetUint64(cliflags.MaxDepth)
	requiredPrefix := viper.GetString(cliflags.RequiredPrefix)
	fmt.Printf("Parse root: %v\n", viper.AllSettings())

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

	ignorePatterns := viper.GetStringSlice(cliflags.IgnorePattern)
	ignoreRegexes := make([]*regexp.Regexp, 0, len(ignorePatterns))
	for _, pattern := range ignorePatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return Config{}, fmt.Errorf("Invalid ignore pattern %s: %v", pattern, err)
		}
		ignoreRegexes = append(ignoreRegexes, regex)
	}

	return Config{
		ParseRoot:           parseRoot,
		OutputDir:           outputDir,
		MaxDepth:            maxDepth,
		RequiredPrefix:      requiredPrefix,
		Environment:         viper.GetString(cliflags.Environment),
		DownloadConcurrency: viper.GetUint32(cliflags.DownloadConcurrency),
		ParseConcurrency:    viper.GetUint32(cliflags.ParseConcurrency),
		IgnorePatterns:      ignoreRegexes,
	}, nil
}
