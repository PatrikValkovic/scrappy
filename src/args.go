package src

import (
	"flag"
	"fmt"
	"os"
)

func printHelp() {
	fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Println("Scrap website and download it locally\n")
	flag.PrintDefaults()
}

type Args struct {
	ParseRoot string
}

func ParseArgs() Args {
	parseRoot := flag.String("parse-root", "", "Where to start parsing")
	help := flag.Bool("help", false, "Print help message")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}
	if *parseRoot == "" {
		fmt.Println("Missing parse root")
		printHelp()
		os.Exit(1)
	}

	return Args{
		ParseRoot: *parseRoot,
	}
}
