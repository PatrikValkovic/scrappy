package parser

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/tdewolff/parse/css"
	"go.uber.org/zap"
)

type CssParser struct {
	Logger   *zap.SugaredLogger
	location url.URL
}

func (this *CssParser) Process(content []byte, arg DownloadArg) ([]byte, []DownloadArg, error) {
	this.location = arg.Url
	parser := css.NewLexer(bytes.NewReader(content))
	out := bytes.NewBuffer([]byte{})
	links := []DownloadArg{}

	for {
		tt, b := parser.Next()
		switch tt {
		case css.ErrorToken:
			err := parser.Err()
			if err == io.EOF {
				result := out.Bytes()
				return result, links, nil
			}
			this.Logger.Errorf("Error parsing css: %s: %v", arg.Url.String(), parser.Err())
			return content, []DownloadArg{}, nil
		case css.URLToken:
			s := string(b)
			link := s[4 : len(s)-1]
			this.Logger.Debugf("Found link in styles: %s", link)
			p := this.handlePath(link, "in-css")
			if !p.success {
				this.Logger.Warnf("Could not parse css link: %s", link)
				out.Write(b)
				continue
			}
			this.Logger.Debugf("Parsed css link %s saved into %s", p.url.String(), p.localPath)
			out.WriteString(fmt.Sprintf("url(\"../%s\")", p.localPath))
			links = append(links, DownloadArg{
				Url:      p.url,
				Depth:    arg.Depth + 1,
				FileName: p.localPath,
			})
		default:
			out.Write(b)
		}
	}
}

func (this *CssParser) handlePath(attr string, localPrefix string) ProcessedPath {
	processor := PathProcessor{
		Logger:   this.Logger,
		Location: this.location,
	}
	return processor.HandlePath(attr, localPrefix)
}
