package parsers

type ImageParser struct {
}

func (this *ImageParser) Process(content *[]byte, _ DownloadArg) (*[]byte, []DownloadArg, error) {
	return content, []DownloadArg{}, nil
}
