package parsers

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type ProcessedPath struct {
	Success     bool
	Url         url.URL
	LocalPath   string
	RelativeUrl string
}

type PathProcessor struct {
	Logger    *zap.SugaredLogger
	Location  *url.URL
	mutex     sync.Mutex
	urlToFile map[string]string
	fileToUrl map[string]string
}

func NewPathProcessor(logger *zap.SugaredLogger, location *url.URL) *PathProcessor {
	return &PathProcessor{
		Logger:    logger,
		Location:  location,
		mutex:     sync.Mutex{},
		urlToFile: make(map[string]string),
		fileToUrl: make(map[string]string),
	}
}

func (this *PathProcessor) HandlePath(
	link string,
	onSite url.URL,
	localPrefix string,
) ProcessedPath {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	fullUrl, err := url.Parse(link)
	if err != nil {
		this.Logger.Warnf("Could not parse css file link: %s", err)
		return ProcessedPath{Success: false}
	}
	resolved := onSite.ResolveReference(fullUrl)
	resolvedWithoutFragment := *resolved
	resolvedWithoutFragment.Fragment = ""

	if existingPath := this.urlToFile[resolvedWithoutFragment.String()]; existingPath != "" {
		return ProcessedPath{
			Success:     true,
			Url:         *resolved,
			LocalPath:   existingPath,
			RelativeUrl: existingPath + resolved.Fragment,
		}
	}

	relativeFileName := resolved.Path
	if strings.HasPrefix(relativeFileName, this.Location.Path) {
		relativeFileName = relativeFileName[len(this.Location.Path):]
	}
	if strings.HasSuffix(relativeFileName, "/index.html") {
		relativeFileName = relativeFileName[:len(relativeFileName)-len("index.html")]
	}
	if strings.HasSuffix(relativeFileName, "/") {
		relativeFileName = relativeFileName[:len(relativeFileName)-1]
	}
	if directory := filepath.Dir(relativeFileName); directory != "." {
		relativeFileName = relativeFileName[len(directory):]
	}
	if relativeFileName == "/" || relativeFileName == "" {
		relativeFileName = "index.html"
	}
	if filepath.Ext(relativeFileName) == "" {
		relativeFileName = relativeFileName + ".html"
	}

	fileName := filepath.Join(localPrefix, relativeFileName)
	counter := 0
	for _, ok := this.fileToUrl[fileName]; ok; _, ok = this.fileToUrl[fileName] {
		counter++
		extension := filepath.Ext(fileName)
		fileName = filepath.Join(
			localPrefix,
			fmt.Sprintf("%s_%d%s", relativeFileName[:len(relativeFileName)-len(extension)], counter, extension),
		)
	}

	this.fileToUrl[fileName] = resolvedWithoutFragment.String()
	this.urlToFile[resolvedWithoutFragment.String()] = fileName

	relativeUrl := fileName
	if resolved.Fragment != "" {
		relativeUrl = fileName + "#" + resolved.Fragment
	}

	return ProcessedPath{
		Success:     true,
		Url:         *resolved,
		LocalPath:   fileName,
		RelativeUrl: relativeUrl,
	}
}
