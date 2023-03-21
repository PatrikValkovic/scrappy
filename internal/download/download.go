package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type DownloadResult struct {
	Url         url.URL
	Content     []byte
	ContentType string
}

func Download(url url.URL, logger *zap.SugaredLogger) (DownloadResult, error) {
	resp, err := http.Get(url.String())
	if resp != nil {
		defer func() {
			closeErr := resp.Body.Close()
			if closeErr != nil {
				logger.Warnf("Error closing response body: %v", closeErr)
			}
		}()
	}
	if err != nil {
		return DownloadResult{}, err
	}
	if resp.StatusCode >= 400 {
		return DownloadResult{}, errors.New(fmt.Sprintf("Download received status %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DownloadResult{}, err
	}
	return DownloadResult{
		Url:         url,
		Content:     body,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}
