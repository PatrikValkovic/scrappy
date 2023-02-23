package parsers

import (
	"net/url"
)

type CssParser struct {
}

func (this *CssParser) Process(content *[]byte, _ url.URL) (*[]byte, []DownloadArg, error) {
	return content, []DownloadArg{}, nil
}
