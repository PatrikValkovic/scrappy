package parser

import (
	"net/url"
	"path/filepath"

	"go.uber.org/zap"
)

type ProcessedPath struct {
	success   bool
	url       url.URL
	localPath string
}

type PathProcessor struct {
	Logger   *zap.SugaredLogger
	Location url.URL
}

func (this *PathProcessor) HandlePath(attr string, localPrefix string) ProcessedPath {
	fullUrl, err := url.Parse(attr)
	if err != nil {
		this.Logger.Warnf("Could not parse css file link: %s", err)
		return ProcessedPath{success: false}
	}
	if fullUrl.Scheme == "" {
		fullUrl.Scheme = this.Location.Scheme
	}
	if fullUrl.Host == "" {
		fullUrl.Host = this.Location.Host
	}
	if fullUrl.Path == "/" {
		fullUrl.Path = "/index.html"
	}
	fileName := filepath.Join(localPrefix, filepath.Base(fullUrl.Path))
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".html"
	}
	return ProcessedPath{
		success:   true,
		url:       *fullUrl,
		localPath: fileName,
	}
}
