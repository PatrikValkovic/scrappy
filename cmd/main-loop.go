package cmd

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/internal/config"
	"github.com/PatrikValkovic/scrappy/internal/download"
	"github.com/PatrikValkovic/scrappy/internal/parsers"
)

func startMainLoop(args *config.Config, logger *zap.SugaredLogger) error {
	logger.Infof("Starting download loop for %s", args.ParseRoot)

	downloadCounter := atomic.Int32{}
	parsingCounter := atomic.Int32{}
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
	downloadCounter.Add(1)
	downloadQueue <- &rootDownloadArg
	processedMutex := sync.Mutex{}
	prcessedSet := make(map[string]interface{})

	// Downloads
	downloadPool := sync.WaitGroup{}
	ctx, ctxCancel := context.WithCancel(context.Background())
	for i := uint32(0); i < args.DownloadConcurrency; i++ {
		downloadPool.Add(1)
		go func() {
			defer downloadPool.Done()
			for downloadCounter.Load() > 0 || parsingCounter.Load() > 0 {
				var downloadArg *parsers.DownloadArg
				select {
				case <-ctx.Done():
					continue
				case downloadArg = <-downloadQueue:
				}
				if downloadArg.Depth > args.MaxDepth {
					logger.Debugf("Skipping %s because of depth", downloadArg.Url.String())
					downloadCounter.Add(-1)
					continue
				}
				processedMutex.Lock()
				if _, ok := prcessedSet[downloadArg.Url.String()]; ok {
					processedMutex.Unlock()
					logger.Debugf("Skipping %s because of already processed", downloadArg.Url.String())
					downloadCounter.Add(-1)
					continue
				}
				prcessedSet[downloadArg.Url.String()] = 1
				processedMutex.Unlock()

				logger.Infof("Downloading %s, remaining: %d", downloadArg.Url.String(), downloadCounter.Load())
				response, err := download.Download(downloadArg.Url, logger)
				if err != nil {
					logger.Warnf("Error downloading %s: %s", downloadArg.Url.String(), err)
					if downloadArg.IsRequired {
						logger.Fatalf("IsRequired download failed")
					}
					downloadCounter.Add(-1)
					continue
				}
				logger.Debugf("Downloaded %s", response.ContentType)

				parseArg := parsers.NewParseArg(*downloadArg, response.Content, response.ContentType)
				parsingCounter.Add(1)
				parseQueue <- &parseArg
				downloadCounter.Add(-1)
			}
			logger.Infoln("Download finished")
			ctxCancel()
		}()
	}

	// Parsers
	parsePool := sync.WaitGroup{}
	for i := uint32(0); i < args.ParseConcurrency; i++ {
		parsePool.Add(1)
		go func() {
			defer parsePool.Done()
			for downloadCounter.Load() > 0 || parsingCounter.Load() > 0 {
				var toParse *parsers.ParseArg
				select {
				case <-ctx.Done():
					continue
				case toParse = <-parseQueue:
				}
				parser := parsers.GetParser(toParse.ContentType, logger, args)
				if parser == nil {
					parsingCounter.Add(-1)
					logger.Warnf("No parser found for %s", toParse.ContentType)
					if toParse.DownloadArg.IsRequired {
						logger.Fatalf("IsRequired download is missing type parser")
					}
					continue
				}

				result, toProcess, err := parser.Process(toParse.Body, toParse.DownloadArg)
				if err != nil {
					parsingCounter.Add(-1)
					logger.Warnf("Error processing %s: %s", toParse.ContentType, err)
					if toParse.DownloadArg.IsRequired {
						logger.Fatalf("IsRequired download failed to process")
					}
					continue
				}
				logger.Infof("Processed %s, returned %d new downloads", toParse.DownloadArg.Url.String(), len(toProcess))
				saveFile(filepath.Join(args.OutputDir, toParse.DownloadArg.FileName), logger, result)
				for _, downloadArg := range toProcess {
					downloadCounter.Add(1)
					downloadQueue <- &downloadArg
				}
				parsingCounter.Add(-1)
			}
			logger.Infoln("Parsing finished")
			ctxCancel()
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
