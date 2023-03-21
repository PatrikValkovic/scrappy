package cmd

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/internal/config"
	"github.com/PatrikValkovic/scrappy/internal/download"
	"github.com/PatrikValkovic/scrappy/internal/parsers"
)

func startMainLoop(args *config.Config, logger *zap.SugaredLogger) error {
	logger.Infof("Starting download loop for %s", args.ParseRoot)

	downloadQueue := make(chan *parsers.DownloadArg, args.DownloadConcurrency)
	parseQueue := make(chan *parsers.ParseArg, args.ParseConcurrency)
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
	downloadQueue <- &rootDownloadArg
	processedMutex := sync.Mutex{}
	prcessedSet := make(map[string]interface{})

	// Downloads
	downloadPool := sync.WaitGroup{}
	for i := uint32(0); i < args.DownloadConcurrency; i++ {
		downloadPool.Add(1)
		go func() {
			defer downloadPool.Done()
			for {
				downloadArg := <-downloadQueue
				if downloadArg.Depth > args.MaxDepth {
					logger.Debugf("Skipping %s because of depth", downloadArg.Url.String())
					continue
				}
				processedMutex.Lock()
				if _, ok := prcessedSet[downloadArg.Url.String()]; ok {
					processedMutex.Unlock()
					continue
				}
				processedMutex.Unlock()

				logger.Infof("Downloading %s, remaining: %d", downloadArg.Url.String(), len(downloadQueue))
				response, err := download.Download(downloadArg.Url, logger)
				if err != nil {
					logger.Warnf("Error downloading %s: %s", downloadArg.Url.String(), err)
					if downloadArg.IsRequired {
						logger.Fatalf("IsRequired download failed")
					}
					continue
				}
				logger.Debugf("Downloaded %s", response.ContentType)

				parseArg := parsers.NewParseArg(*downloadArg, response.Content, response.ContentType)
				parseQueue <- &parseArg
			}
		}()
	}

	// Parsers
	parsePool := sync.WaitGroup{}
	for i := uint32(0); i < args.ParseConcurrency; i++ {
		parsePool.Add(1)
		go func() {
			defer parsePool.Done()
			for {
				toParse := <-parseQueue

				parser := parsers.GetParser(toParse.ContentType, logger, args)
				if parser == nil {
					logger.Warnf("No parser found for %s", toParse.ContentType)
					if toParse.DownloadArg.IsRequired {
						logger.Fatalf("IsRequired download is missing type parser")
					}
					continue
				}

				result, toProcess, err := parser.Process(toParse.Body, toParse.DownloadArg)
				if err != nil {
					logger.Warnf("Error processing %s: %s", toParse.ContentType, err)
					if toParse.DownloadArg.IsRequired {
						logger.Fatalf("IsRequired download failed to process")
					}
					continue
				}
				logger.Infof("Processed %s, returned %d new downloads", toParse.DownloadArg.Url.String(), len(toProcess))
				saveFile(filepath.Join(args.OutputDir, toParse.DownloadArg.FileName), logger, result)
				for _, downloadArg := range toProcess {
					downloadQueue <- &downloadArg
				}
			}
		}()
	}

	parsePool.Wait()
	downloadPool.Wait()
	return nil
}

func saveFile(path string, logger *zap.SugaredLogger, content []byte) {
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
	written, err := file.Write(content)
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
