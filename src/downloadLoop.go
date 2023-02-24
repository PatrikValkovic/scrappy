package src

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/src/args"
	"github.com/PatrikValkovic/scrappy/src/parsers"
)

func Start(args *args.Args, logger *zap.SugaredLogger) {
	logger.Infof("Starting download loop for %s", args.ParseRoot)

	downloadQueue := make([]parsers.DownloadArg, 0)
	rootDownloadArg, rootDownloadError := parsers.NewDownloadArg(
		args.ParseRoot,
		true,
		"index.html",
		logger,
		0,
	)
	if rootDownloadError != nil {
		logger.Fatalf("Could not parse root url %s", args.ParseRoot)
	}
	downloadQueue = append(downloadQueue, rootDownloadArg)
	prcessedSet := make(map[string]interface{})

	for len(downloadQueue) > 0 {
		downloadArg := downloadQueue[0]
		downloadQueue = downloadQueue[1:]
		logger.Infof("Downloading %s, remaining: %d", downloadArg.Url.String(), len(downloadQueue))

		if downloadArg.Depth > args.MaxDepth {
			logger.Debugf("Skipping %s because of depth", downloadArg.Url.String())
			continue
		}
		if _, ok := prcessedSet[downloadArg.Url.String()]; ok {
			continue
		}
		prcessedSet[downloadArg.Url.String()] = byte(0)

		response, err := Download(downloadArg.Url, logger)
		if err != nil {
			logger.Warnf("Error downloading %s: %s", downloadArg.Url.String(), err)
			if downloadArg.IsRequired {
				logger.Fatalf("IsRequired download failed")
			}
			continue
		}

		logger.Debugf("Downloaded %s", response.ContentType)
		parser := parsers.GetParser(response.ContentType, logger, args)
		if parser == nil {
			logger.Warnf("No parser found for %s", response.ContentType)
			if downloadArg.IsRequired {
				logger.Fatalf("IsRequired download is missing type parser")
			}
			continue
		}

		result, toProcess, err := parser.Process(response.Content, downloadArg)
		if err != nil {
			logger.Warnf("Error processing %s: %s", response.ContentType, err)
			if downloadArg.IsRequired {
				logger.Fatalf("IsRequired download failed to process")
			}
			continue
		}
		logger.Infof("Processed %s, returned %d new downloads", response.Url.String(), len(toProcess))
		downloadQueue = append(downloadQueue, toProcess...)

		saveFile(filepath.Join(args.OutputDir, downloadArg.FileName), logger, result)
	}
}

func saveFile(path string, logger *zap.SugaredLogger, content *[]byte) {
	outputDir := filepath.Dir(path)
	_, err := os.ReadDir(outputDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			logger.Fatalf("Could not create output directory %s because of %v", outputDir, err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		logger.Fatalf("Could not create file %s because of %v", path, err)
	}
	written, err := file.Write(*content)
	logger.Debugf("Written %d bytes", written)
	if err != nil {
		logger.Fatalf("Could not write to file %s because of %v", path, err)
	}

	err = file.Close()
	if err != nil {
		logger.Fatalf("Error closing file %s because of %v", path, err)
	}

	logger.Debugf("Data written into %s", path)
}
