package main

import (
	"github.com/PatrikValkovic/scrappy/src"
	"github.com/PatrikValkovic/scrappy/src/arg"
)

func main() {
	logger := src.CreateLogger()
	src.LoadEnvironment(logger)
	logger = src.CreateLogger()

	args := arg.ParseArgs()
	src.Start(&args, logger)
}
