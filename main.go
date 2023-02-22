package main

import src "github.com/PatrikValkovic/scrappy/src"

func main() {
	logger := src.CreateLogger()
	src.LoadEnvironment(logger)
	logger = src.CreateLogger()

	args := src.ParseArgs()
	src.Start(&args, logger)
}
