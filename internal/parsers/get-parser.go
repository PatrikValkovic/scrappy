package parsers

import (
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/internal/config"
)

type Parser interface {
	Process(content []byte, download DownloadArg) ([]byte, []DownloadArg, error)
}

func GetParser(
	contentType string,
	logger *zap.SugaredLogger,
	args *config.Config,
) Parser {
	switch true {
	case strings.Contains(contentType, "text/html"):
		return &HtmlParser{Logger: logger, Args: args}
	case strings.Contains(contentType, "text/css"):
		return &CssParser{Logger: logger}
	case strings.Contains(contentType, "javascript"):
		return &JavaScriptParser{}
	case strings.HasPrefix(contentType, "image/"):
		return &PassthroughParser{}
	case strings.HasPrefix(contentType, "font/"):
		return &PassthroughParser{}
	case strings.HasPrefix(contentType, "video/"):
		return &PassthroughParser{}
	default:
		return nil
	}
}

type DownloadArg struct {
	Url        url.URL
	IsRequired bool
	FileName   string
	Depth      uint64
}

type ParseArg struct {
	DownloadArg DownloadArg
	Body        []byte
	ContentType string
}

func NewDownloadArg(
	link string,
	required bool,
	fileName string,
	logger *zap.SugaredLogger,
	depth uint64,
) (DownloadArg, error) {
	parsedUrl, err := url.Parse(link)
	if err != nil && required {
		logger.Fatalf("Url \"%s\" is not a valid URL", link)
	}
	if err != nil {
		return DownloadArg{}, err
	}
	return DownloadArg{
		Url:        *parsedUrl,
		IsRequired: required,
		FileName:   fileName,
		Depth:      depth,
	}, nil
}

func NewParseArg(
	downloadArg DownloadArg,
	body []byte,
	contentType string,
) ParseArg {
	return ParseArg{
		DownloadArg: downloadArg,
		Body:        body,
		ContentType: contentType,
	}
}
