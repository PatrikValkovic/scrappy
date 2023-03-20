package main

import (
	"os"

	"github.com/PatrikValkovic/scrappy/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
