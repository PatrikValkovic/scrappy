package arg

import (
	"flag"
	"fmt"
	"os"
)

func printHelp() {
	fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Printf("Scrap website and download it locally\n\n")
	flag.PrintDefaults()
}

type Args struct {
	ParseRoot      string
	OutputDir      string
	MaxDepth       uint64
	RequiredPrefix string
}

func ParseArgs() Args {
	parseRoot := flag.String("parse-root", "", "Where to start parsing")
	outputDir := flag.String("output-dir", "", "Where to store downloaded files")
	maxDepth := flag.Uint64("max-depth", 20, "Where to store downloaded files")
	requiredPrefix := flag.String("required-prefix", "", "Prefix that all the links must have")
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
	if *outputDir == "" {
		fmt.Println("Missing output dir")
		printHelp()
		os.Exit(1)
	}
	if *requiredPrefix == "" {
		*requiredPrefix = *parseRoot
	}

	return Args{
		ParseRoot:      *parseRoot,
		OutputDir:      *outputDir,
		MaxDepth:       *maxDepth,
		RequiredPrefix: *requiredPrefix,
	}
}
