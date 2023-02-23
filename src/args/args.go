package args

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
	OutputDir string
}

func ParseArgs() Args {
	parseRoot := flag.String("parse-root", "", "Where to start parsing")
	outputDir := flag.String("output-dir", "", "Where to store downloaded files")
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

	return Args{
		ParseRoot: *parseRoot,
		OutputDir: *outputDir,
	}
}
