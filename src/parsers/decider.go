package parsers

import (
	"net/url"

	"go.uber.org/zap"

	"github.com/PatrikValkovic/scrappy/src/args"
)

type Parser interface {
	Process(content *[]byte, location url.URL) (*[]byte, []DownloadArg, error)
}

func GetParser(
	contentType string,
	logger *zap.SugaredLogger,
	args *args.Args,
) Parser {
	switch true {
	case contentType == "text/html":
		return &HtmlParser{Logger: logger, Args: args}
	case contentType == "text/css":
		return &CssParser{}
	default:
		return nil
	}
}

type DownloadArg struct {
	Url        url.URL
	IsRequired bool
	FileName   string
}

func NewDownloadArg(
	link string,
	required bool,
	fileName string,
	logger *zap.SugaredLogger,
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
	}, nil
}
