package cmd

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/adrianbrad/queue"
	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/internal/config"
	"github.com/PatrikValkovic/scrappy/internal/download"
	"github.com/PatrikValkovic/scrappy/internal/parsers"
)

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func startMainLoop(args *config.Config, logger *zap.SugaredLogger) error {
	var err error
	logger.Infof("Starting download loop for %s", args.ParseRoot)

	downloadCoordChannel := make(chan uint32)
	parserCoordChannel := make(chan uint32)
	downloadQueue := queue.NewLinked([]parsers.DownloadArg{})
	parseQueue := queue.NewBlocking([]*parsers.ParseArg{}, queue.WithCapacity(4*int(args.ParseConcurrency)))
	interruptCtx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	endCtx, endProgram := context.WithCancel(context.Background())
	prefixUrl, err := url.Parse(args.RequiredPrefix)
	if err != nil {
		logger.Fatalf("Could not parse prefix url %s", args.ParseRoot)
	}
	pathProcessor := parsers.NewPathProcessor(logger, prefixUrl)

	processedMutex := sync.Mutex{}
	prcessedSet := make(map[string]interface{})

	// Root download
	processedRoot := pathProcessor.HandlePath(args.ParseRoot, *prefixUrl, ".")
	if !processedRoot.Success {
		logger.Fatalf("Could not parse root url %s", args.ParseRoot)
	}
	rootDownloadArg, rootDownloadError := parsers.NewDownloadArg(
		args.ParseRoot,
		true,
		processedRoot.LocalPath,
		logger,
		0,
	)
	if rootDownloadError != nil {
		logger.Fatalf("Could not parse root url %s", args.ParseRoot)
	}
	err = downloadQueue.Offer(rootDownloadArg)
	if err != nil {
		logger.Fatalf("Could not insert root url %s into queue", args.ParseRoot)
	}

	// Coordinator
	go func() {
		exitingDownloaders := []uint32{}
		exitingParsers := []uint32{}
		for len(exitingParsers) < int(args.ParseConcurrency) ||
			len(exitingDownloaders) < int(args.DownloadConcurrency) ||
			!downloadQueue.IsEmpty() ||
			!parseQueue.IsEmpty() {

			select {
			case <-time.After(100 * time.Millisecond):
				if !downloadQueue.IsEmpty() || !parseQueue.IsEmpty() {
					exitingDownloaders = exitingDownloaders[:0]
					exitingParsers = exitingParsers[:0]
				}
				continue
			case <-interruptCtx.Done():
				logger.Debugln("Coordinator interrupted")
				return
			case downloderIdentifier := <-downloadCoordChannel:
				if !downloadQueue.IsEmpty() || !parseQueue.IsEmpty() {
					continue
				}
				if Contains(exitingDownloaders, downloderIdentifier) {
					continue
				}
				exitingDownloaders = append(exitingDownloaders, downloderIdentifier)
			case paserIdentifier := <-parserCoordChannel:
				if !downloadQueue.IsEmpty() || !parseQueue.IsEmpty() {
					continue
				}
				if Contains(exitingParsers, paserIdentifier) {
					continue
				}
				exitingParsers = append(exitingParsers, paserIdentifier)
			}
		}

		logger.Infoln("All downloaders and parsers finished")
		endProgram()
	}()

	// Downloads
	downloadPool := sync.WaitGroup{}
	for i := uint32(0); i < args.DownloadConcurrency; i++ {
		downloadPool.Add(1)
		go func(identifier uint32) {
			defer downloadPool.Done()
			for true {
				var downloadArg parsers.DownloadArg
				select {
				case <-endCtx.Done():
					return
				case <-interruptCtx.Done():
					logger.Debugln("Downloader interrupted")
					return
				default:
					if downloadQueue.IsEmpty() {
						downloadCoordChannel <- identifier
						time.Sleep(100 * time.Millisecond)
						continue
					}
					downloadArg, err = downloadQueue.Get()
					if err != nil {
						logger.Warnf("Error getting download from queue: %s", err)
						continue
					}
				}

				if downloadArg.Depth > args.MaxDepth {
					logger.Debugf("Skipping %s because of depth", downloadArg.Url.String())
					continue
				}
				processedMutex.Lock()
				if _, ok := prcessedSet[downloadArg.Url.String()]; ok {
					processedMutex.Unlock()
					logger.Debugf("Skipping %s because of already processed", downloadArg.Url.String())
					continue
				}
				prcessedSet[downloadArg.Url.String()] = 1
				processedMutex.Unlock()

				logger.Infof("Downloading %s, remaining: %d", downloadArg.Url.String(), downloadQueue.Size())
				response, err := download.Download(downloadArg.Url, logger)
				if err != nil {
					logger.Warnf("Error downloading %s: %s", downloadArg.Url.String(), err)
					if downloadArg.IsRequired {
						logger.Fatalf("IsRequired download failed")
					}
					continue
				}
				logger.Debugf("Downloaded %s", response.ContentType)

				parseArg := parsers.NewParseArg(downloadArg, response.Content, response.ContentType)
				parseQueue.OfferWait(&parseArg)
			}
			logger.Infoln("Download finished")
			endProgram()
		}(i)
	}

	// Parsers
	parsePool := sync.WaitGroup{}
	for i := uint32(0); i < args.ParseConcurrency; i++ {
		parsePool.Add(1)
		go func(identifier uint32) {
			defer parsePool.Done()
			for true {
				var toParse *parsers.ParseArg
				select {
				case <-endCtx.Done():
					return
				case <-interruptCtx.Done():
					logger.Debugln("Parser interrupted")
					return
				default:
					if parseQueue.IsEmpty() {
						parserCoordChannel <- identifier
						time.Sleep(100 * time.Millisecond)
						continue
					}
					toParse, err = parseQueue.Get()
					if err != nil {
						logger.Warnf("Error getting parse from queue: %s", err)
						continue
					}
				}
				parser := parsers.GetParser(toParse.ContentType, logger, args, pathProcessor)
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
					tmp := downloadArg
					err := downloadQueue.Offer(tmp)
					if err != nil {
						logger.Warnf("Error inserting download into queue: %s", err)
					}
				}
			}
			logger.Infoln("Parsing finished")
			endProgram()
		}(i)
	}

	parsePool.Wait()
	downloadPool.Wait()
	endProgram()
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
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Fatalf("Error closing file %s because of %v", path, err)
		}
	}()
	written, err := file.Write(content)
	logger.Debugf("Written %d bytes", written)
	if err != nil {
		logger.Fatalf("Could not write to file %s because of %v", path, err)
	}

	logger.Debugf("Data written into %s", path)
}
