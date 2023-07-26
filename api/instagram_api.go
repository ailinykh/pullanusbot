package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateInstagramAPI
func CreateInstagramAPI(l core.ILogger, cookie string) InstAPI {
	return &InstagramAPI{l, cookie}
}

type InstAPI interface {
	GetReel(string) (*IgReel, error)
}

// Instagram API
type InstagramAPI struct {
	l      core.ILogger
	cookie string
}

func (api *InstagramAPI) GetReel(urlString string) (*IgReel, error) {
	body, err := api.getContent(urlString, map[string]string{
		"sec-fetch-mode": "navigate",
		"user-agent":     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
		"cookie":         api.cookie,
		"accept":         "text/html",
	})
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	// os.WriteFile("instagram-reel.html", body, 0644)

	r := regexp.MustCompile(`"xdt_api__v1__media__shortcode__web_info":(.*)\},"extensions"`)
	match := r.FindSubmatch(body)
	if len(match) < 2 {
		return nil, fmt.Errorf("parse HTML failed: %s", urlString)
	}

	// os.WriteFile("instagram-reel.json", match[1], 0644)

	var reel IgReel
	err = json.Unmarshal(match[1], &reel)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	return &reel, nil
}

func (api *InstagramAPI) getContent(urlString string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", urlString, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
