package parsers

import (
	"bufio"
	"bytes"
	"errors"
	"net/url"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/src/args"
)

type HtmlParser struct {
	Logger *zap.SugaredLogger
	Args   *args.Args

	location url.URL
}

func (this *HtmlParser) Process(content *[]byte, location url.URL) (*[]byte, []DownloadArg, error) {
	this.location = location

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(*content))
	if err != nil {
		return nil, nil, errors.New("Could not parse html")
	}

	cssDownloads := this.processCss(document)
	imageDownloads := this.processImages(document)

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	err = goquery.Render(writer, document.Selection)
	if err != nil {
		return nil, nil, errors.New("Could not render html")
	}
	err = writer.Flush()
	if err != nil {
		return nil, nil, errors.New("Could not flush html into buffer")
	}

	result := buffer.Bytes()
	return &result, concat([][]DownloadArg{cssDownloads, imageDownloads}), nil
}

func concat(slices [][]DownloadArg) []DownloadArg {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]DownloadArg, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}

func (this *HtmlParser) processCss(document *goquery.Document) []DownloadArg {
	cssFiles := document.Find("link[rel=\"stylesheet\"]")
	cssDownloads := make([]DownloadArg, 0)
	if cssFiles == nil {
		this.Logger.Debugf("No css files found")
	} else {
		this.Logger.Debugf("Found %d css files", cssFiles.Length())
		cssFiles.Each(func(i int, s *goquery.Selection) {
			hrefAttr := s.AttrOr("href", "")
			this.Logger.Debugf("Found css file: %s", hrefAttr)
			processed := this.handlePath(hrefAttr, "styles")
			if !processed.success {
				this.Logger.Warnf("Could not parse css file link: %s", hrefAttr)
				return
			}
			this.Logger.Debugf("Style %s will be stored into %s", processed.url.String(), processed.localPath)
			downloadArg, err := NewDownloadArg(
				processed.url.String(),
				true,
				processed.localPath,
				this.Logger,
			)
			s.SetAttr("href", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not parse css file link: %s", err)
				return
			}
			cssDownloads = append(cssDownloads, downloadArg)
		})
	}
	return cssDownloads
}

func (this *HtmlParser) processImages(document *goquery.Document) []DownloadArg {
	imgElements := document.Find("img[src]")
	imgDownloads := make([]DownloadArg, 0)
	if imgElements == nil {
		this.Logger.Debugf("No image files found")
	} else {
		this.Logger.Debugf("Found %d img files", imgElements.Length())
		imgElements.Each(func(i int, s *goquery.Selection) {
			srcAttr := s.AttrOr("src", "")
			this.Logger.Debugf("Found img file: %s", srcAttr)
			processed := this.handlePath(srcAttr, "img")
			if !processed.success {
				this.Logger.Warnf("Could not parse img file link: %s", srcAttr)
				return
			}
			this.Logger.Debugf("Image %s will be stored into %s", processed.url.String(), processed.localPath)
			downloadArg, err := NewDownloadArg(
				processed.url.String(),
				false,
				processed.localPath,
				this.Logger,
			)
			s.SetAttr("src", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not parse img file link: %s", err)
				return
			}
			imgDownloads = append(imgDownloads, downloadArg)
		})
	}
	return imgDownloads
}

type processedPath struct {
	success   bool
	url       url.URL
	localPath string
}

func (this *HtmlParser) handlePath(attr string, localPrefix string) processedPath {
	fullUrl, err := url.Parse(attr)
	if err != nil {
		this.Logger.Warnf("Could not parse css file link: %s", err)
		return processedPath{success: false}
	}
	if fullUrl.Scheme == "" {
		fullUrl.Scheme = this.location.Scheme
	}
	if fullUrl.Host == "" {
		fullUrl.Host = this.location.Host
	}
	fileName := filepath.Join(localPrefix, filepath.Base(fullUrl.Path))
	return processedPath{
		success:   true,
		url:       *fullUrl,
		localPath: fileName,
	}
}
