package src

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func LoadEnvironment(logger *zap.SugaredLogger) {
	err := godotenv.Load()
	if os.IsNotExist(err) {
		logger.Warn("Env file not found")
		return
	}
	if err != nil {
		logger.Fatal("Error loading env file: %v", err)
	}
}
