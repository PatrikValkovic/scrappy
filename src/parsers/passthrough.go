package parsers

type PassthroughParser struct {
}

func (this *PassthroughParser) Process(content *[]byte, _ DownloadArg) (*[]byte, []DownloadArg, error) {
	return content, []DownloadArg{}, nil
}
