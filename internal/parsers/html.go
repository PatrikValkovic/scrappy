package parsers

import (
	"bufio"
	"bytes"
	"errors"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/internal/args"
)

type HtmlParser struct {
	Logger *zap.SugaredLogger
	Args   *args.Args

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
	document.Find("noscript").Each(func(i int, s *goquery.Selection) {
		s.ReplaceWithHtml(s.Text())
	})

	cssDownloads := this.processCss(document)
	imageDownloads := this.processImages(document)
	linksDownloads := this.processLinks(document)
	scriptsDownloads := this.processScripts(document)
	videoDownloads := this.processVideo(document)

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
		videoDownloads,
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

func (this *HtmlParser) processVideo(document *goquery.Document) []DownloadArg {
	videoDownloads := make([]DownloadArg, 0)

	videoElements := document.Find("video")
	if videoElements == nil {
		this.Logger.Debugf("No video files found")
	} else {
		this.Logger.Debugf("Found %d video files", videoElements.Length())
		videoElements.Each(func(i int, s *goquery.Selection) {
			attr := s.AttrOr("poster", "")
			if attr == "" {
				this.Logger.Debugf("Skipping video without poster")
				return
			}
			this.Logger.Debugf("Found video poster: %s", attr)
			processed := this.handlePath(attr, "video")
			if !processed.success {
				this.Logger.Warnf("Could not parse video poster link: %s", attr)
				return
			}
			this.Logger.Debugf("Video poster %s will be stored into %s", processed.url.String(), processed.localPath)
			downloadArg, err := NewDownloadArg(
				processed.url.String(),
				false,
				processed.localPath,
				this.Logger,
				this.depth,
			)
			s.SetAttr("poster", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not create video poster link: %s", err)
				return
			}
			videoDownloads = append(videoDownloads, downloadArg)
		})
	}

	sourceElements := document.Find("source")
	if sourceElements == nil {
		this.Logger.Debugf("No video sources found")
	} else {
		this.Logger.Debugf("Found %d video sources", sourceElements.Length())
		sourceElements.Each(func(i int, s *goquery.Selection) {
			attr := s.AttrOr("src", "")
			if attr == "" {
				this.Logger.Debugf("Video source not found")
				return
			}
			this.Logger.Debugf("Found video source: %s", attr)
			processed := this.handlePath(attr, "video")
			if !processed.success {
				this.Logger.Warnf("Could not parse video source link: %s", attr)
				return
			}
			this.Logger.Debugf("Video source %s will be stored into %s", processed.url.String(), processed.localPath)
			downloadArg, err := NewDownloadArg(
				processed.url.String(),
				false,
				processed.localPath,
				this.Logger,
				this.depth,
			)
			s.SetAttr("src", processed.localPath)
			if err != nil {
				this.Logger.Warnf("Could not create video source link: %s", err)
				return
			}
			videoDownloads = append(videoDownloads, downloadArg)
		})
	}

	return videoDownloads
}

func (this *HtmlParser) handlePath(attr string, localPrefix string) ProcessedPath {
	processor := PathProcessor{
		Logger:   this.Logger,
		Location: this.location,
	}
	return processor.HandlePath(attr, localPrefix)
}
