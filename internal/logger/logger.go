package logger

import (
	"log"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/PatrikValkovic/scrappy/internal/cliflags"
	"github.com/PatrikValkovic/scrappy/internal/environment"
)

func CreateLogger() *zap.SugaredLogger {
	var config zap.Config
	if viper.GetString(cliflags.Environment) == environment.Development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	config.OutputPaths = []string{"stdout"}

	logger, err := config.Build()
	if err != nil {
		log.Fatal("Can't create logger")
	}
	return logger.Sugar()
}
