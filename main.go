package main

import src "github.com/PatrikValkovic/scrappy/src"

func main() {
	logger := src.CreateLogger()
	src.LoadEnvironment(logger)
	logger = src.CreateLogger()

	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")
	logger.Fatal("Fatal message")
}
