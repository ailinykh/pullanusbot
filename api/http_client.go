package api

import (
	"io/ioutil"
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
	defer resp.Body.Close()

	return resp.Header["Content-Type"][0], nil
}

// GetContent is a core.IHttpClient interface implementation
func (HttpClient) GetContent(url core.URL) (string, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:59.0) Gecko/20100101 Firefox/59.0")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
