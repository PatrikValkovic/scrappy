package main

import (
	src "github.com/PatrikValkovic/scrappy/src"
	args2 "github.com/PatrikValkovic/scrappy/src/args"
)

func main() {
	logger := src.CreateLogger()
	src.LoadEnvironment(logger)
	logger = src.CreateLogger()

	args := args2.ParseArgs()
	src.Start(&args, logger)
}
