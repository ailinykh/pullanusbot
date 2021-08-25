package api

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateHttpClient() *HttpClient {
	return &HttpClient{}
}

type HttpClient struct{}

func (HttpClient) GetRedirectLocation(url core.URL) (core.URL, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:59.0) Gecko/20100101 Firefox/59.0")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	return res.Request.URL.String(), nil
}

// GetContentType is a core.IHttpClient interface implementation
func (HttpClient) GetContentType(url core.URL) (string, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:59.0) Gecko/20100101 Firefox/59.0")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if header, ok := res.Header["Content-Type"]; ok {
		return header[0], nil
	}
	return "", errors.New("content-type not found")
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
