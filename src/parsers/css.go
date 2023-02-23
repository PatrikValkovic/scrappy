package parsers

import (
	"net/url"
)

type CssParser struct {
}

func (this *CssParser) Process(content *[]byte, location url.URL) (*[]byte, []DownloadArg, error) {
	return content, []DownloadArg{}, nil
}
