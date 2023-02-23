package parsers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
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

	cssDownloads := this.ProcessCss(document)

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
	return &result, append(
		[]DownloadArg{},
		cssDownloads...,
	), nil
}

func (this *HtmlParser) ProcessCss(document *goquery.Document) []DownloadArg {
	cssFiles := document.Find("link[rel=\"stylesheet\"]")
	cssDownloads := make([]DownloadArg, 0)
	if cssFiles == nil {
		this.Logger.Debugf("No css files found")
	} else {
		this.Logger.Debugf("Found %d css files", cssFiles.Length())
		cssFiles.Each(func(i int, s *goquery.Selection) {
			absolutePath, err := url.JoinPath(
				fmt.Sprintf("%s://%s", this.location.Scheme, this.location.Host),
				s.AttrOr("href", ""),
			)
			// TODO deterministic file name
			fileName := fmt.Sprintf("%s.css", uuid.New())
			if err != nil {
				this.Logger.Warnf("Could not parse css file link: %s", err)
				return
			}
			downloadArg, err := NewDownloadArg(
				absolutePath,
				true,
				fileName,
				this.Logger,
			)
			s.SetAttr("href", fileName)
			if err != nil {
				this.Logger.Warnf("Could not parse css file link: %s", err)
				return
			}
			cssDownloads = append(cssDownloads, downloadArg)
		})
	}
	return cssDownloads
}
