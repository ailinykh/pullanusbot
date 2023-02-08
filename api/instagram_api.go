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
func CreateInstagramAPI(l core.ILogger, jar http.CookieJar) InstAPI {
	client := http.Client{
		Jar: jar,
	}
	return &InstagramAPI{l, client}
}

type InstAPI interface {
	GetReel(string) (*IgReel, error)
}

type InstagramHTMLData struct {
	appId     string
	csrfToken string
	mediaId   string
}

// Instagram API
type InstagramAPI struct {
	l      core.ILogger
	client http.Client
}

func (api *InstagramAPI) GetReel(urlString string) (*IgReel, error) {
	body, err := api.getContent(urlString, map[string]string{"sec-fetch-mode": "navigate"})
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	// os.WriteFile("instagram-reel.html", body, 0644)

	data, err := api.parseData(body)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	urlString = "https://i.instagram.com/api/v1/media/" + data.mediaId + "/info/"
	body, err = api.getContent(urlString, map[string]string{"x-ig-app-id": data.appId})
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	// os.WriteFile("instagram-reel-"+data.mediaId+".json", body, 0644)

	var reel IgReel
	err = json.Unmarshal(body, &reel)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	return &reel, nil
}

func (api *InstagramAPI) getContent(urlString string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", urlString, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (p *InstagramAPI) parseData(data []byte) (*InstagramHTMLData, error) {
	appId, err := p.parse(data, `"app_id":"(\d+)"`)
	if err != nil {
		return nil, err
	}

	csrfToken, err := p.parse(data, `"csrf_token":"(\w+)"`)
	if err != nil {
		return nil, err
	}

	mediaId, err := p.parse(data, `"media_id":"(\d+)"`)
	if err != nil {
		return nil, err
	}

	return &InstagramHTMLData{string(appId), string(csrfToken), string(mediaId)}, nil
}

func (p *InstagramAPI) parse(data []byte, reg string) ([]byte, error) {
	r := regexp.MustCompile(reg)
	match := r.FindSubmatch(data)
	if len(match) < 2 {
		return nil, fmt.Errorf("parse `%s` failed", reg)
	}
	return match[1], nil
}
