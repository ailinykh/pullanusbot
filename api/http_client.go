package api

import (
	"net/http"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateHttpClient() *HttpClient {
	return &HttpClient{}
}

type HttpClient struct{}

// GetContentType is a core.IHttpClient interface implementation
func (HttpClient) GetContentType(url core.URL) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	return resp.Header["Content-Type"][0], nil
}
