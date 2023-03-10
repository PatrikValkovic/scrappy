package parser

import (
	"bufio"
	"bytes"
	"errors"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/src/arg"
)

type HtmlParser struct {
	Logger *zap.SugaredLogger
	Args   *arg.Args

	location url.URL
	depth    uint64
}

func (this *HtmlParser) Process(content []byte, download DownloadArg) ([]byte, []DownloadArg, error) {
	this.location = download.Url
	this.depth = download.Depth

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, nil, errors.New("Could not parse html")
	}

	cssDownloads := this.processCss(document)
	imageDownloads := this.processImages(document)
	linksDownloads := this.processLinks(document)
	scriptsDownloads := this.processScripts(document)

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
	return result, concat([][]DownloadArg{
		cssDownloads,
		imageDownloads,
		linksDownloads,
		scriptsDownloads,
	}), nil
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
				false,
				processed.localPath,
				this.Logger,
				this.depth,
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
			if strings.HasPrefix(srcAttr, "data:") {
				this.Logger.Debugf("Skipping inline image")
				return
			}
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
				this.depth,
			)
			s.SetAttr("src", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not create image download link: %s", err)
				return
			}
			imgDownloads = append(imgDownloads, downloadArg)
		})
	}
	return imgDownloads
}

func (this *HtmlParser) processScripts(document *goquery.Document) []DownloadArg {
	scriptsElements := document.Find("script[src]")
	scriptDownloads := make([]DownloadArg, 0)
	if scriptsElements == nil {
		this.Logger.Debugf("No scripts files found")
	} else {
		this.Logger.Debugf("Found %d script files", scriptsElements.Length())
		scriptsElements.Each(func(i int, s *goquery.Selection) {
			srcAttr := s.AttrOr("src", "")
			if srcAttr == "" {
				this.Logger.Debugf("Skipping inline script")
				return
			}
			this.Logger.Debugf("Found script file: %s", srcAttr)
			processed := this.handlePath(srcAttr, "js")
			if !processed.success {
				this.Logger.Warnf("Could not parse script file link: %s", srcAttr)
				return
			}
			this.Logger.Debugf("Script %s will be stored into %s", processed.url.String(), processed.localPath)
			downloadArg, err := NewDownloadArg(
				processed.url.String(),
				false,
				processed.localPath,
				this.Logger,
				this.depth,
			)
			s.SetAttr("src", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not create script download link: %s", err)
				return
			}
			scriptDownloads = append(scriptDownloads, downloadArg)
		})
	}
	return scriptDownloads
}

func (this *HtmlParser) processLinks(document *goquery.Document) []DownloadArg {
	links := document.Find("a[href]")
	linksDownloads := make([]DownloadArg, 0)
	if links == nil {
		this.Logger.Debugf("No links found")
	} else {
		this.Logger.Debugf("Found %d links", links.Length())
		links.Each(func(i int, s *goquery.Selection) {
			hrefAttr := s.AttrOr("href", "")
			this.Logger.Debugf("Found link %s", hrefAttr)
			processed := this.handlePath(hrefAttr, ".")
			if !processed.success {
				this.Logger.Warnf("Could not parse link href: %s", hrefAttr)
				return
			}
			if !strings.HasPrefix(processed.url.String(), this.Args.RequiredPrefix) {
				this.Logger.Debugf("Link %s does not have required prefix %s", processed.url.String(), this.Args.RequiredPrefix)
				return
			}
			this.Logger.Debugf("Link %s will be stored into %s", processed.url.String(), processed.localPath)
			downloadArg, err := NewDownloadArg(
				processed.url.String(),
				false,
				processed.localPath,
				this.Logger,
				this.depth+1,
			)
			s.SetAttr("href", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not parse link href: %s", err)
				return
			}
			linksDownloads = append(linksDownloads, downloadArg)
		})
	}
	return linksDownloads
}

func (this *HtmlParser) handlePath(attr string, localPrefix string) ProcessedPath {
	processor := PathProcessor{
		Logger:   this.Logger,
		Location: this.location,
	}
	return processor.HandlePath(attr, localPrefix)
}
