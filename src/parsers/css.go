package parsers

type CssParser struct {
}

func (this *CssParser) Process(content *[]byte, _ DownloadArg) (*[]byte, []DownloadArg, error) {
	return content, []DownloadArg{}, nil
}
