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
	Success   bool
	Url       url.URL
	LocalPath string
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

	if existingPath := this.urlToFile[resolved.String()]; existingPath != "" {
		return ProcessedPath{
			Success:   true,
			Url:       *resolved,
			LocalPath: existingPath,
		}
	}

	clonedUrl := *resolved
	if strings.HasPrefix(clonedUrl.Path, this.Location.Path) {
		clonedUrl.Path = clonedUrl.Path[len(this.Location.Path):]
	}
	if clonedUrl.Path == "/" || clonedUrl.Path == "" {
		clonedUrl.Path = "/index.html"
	}
	fileName := filepath.Join(localPrefix, clonedUrl.Path)
	if strings.HasSuffix(fileName, "/") {
		fileName = fileName[:len(fileName)-1]
	}
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".html"
	}

	counter := 0
	for _, ok := this.fileToUrl[fileName]; ok; _, ok = this.fileToUrl[fileName] {
		counter++
		extension := filepath.Ext(fileName)
		fileName = fmt.Sprintf("%s_%d%s", fileName[:len(fileName)-len(extension)], counter, extension)
	}

	this.fileToUrl[fileName] = resolved.String()
	this.urlToFile[resolved.String()] = fileName

	return ProcessedPath{
		Success:   true,
		Url:       *resolved,
		LocalPath: fileName,
	}
}
