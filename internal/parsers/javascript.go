package parsers

type JavaScriptParser struct {
}

func (this *JavaScriptParser) Process(content []byte, _ DownloadArg) ([]byte, []DownloadArg, error) {
	return content, nil, nil
}
