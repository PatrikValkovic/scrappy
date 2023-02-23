package parsers

import (
	"net/url"
)

type ImageParser struct {
}

func (this *ImageParser) Process(content *[]byte, _ url.URL) (*[]byte, []DownloadArg, error) {
	return content, []DownloadArg{}, nil
}
