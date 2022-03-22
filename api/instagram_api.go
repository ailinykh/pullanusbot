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
func CreateInstagramAPI(l core.ILogger, cookiesFile string) *InstagramAPI {
	jar := CreateCookieJar(l, cookiesFile)
	client := http.Client{
		Jar: jar,
	}

	return &InstagramAPI{l, client}
}

// Instagram API
type InstagramAPI struct {
	l      core.ILogger
	client http.Client
}

func (api *InstagramAPI) GetReel(url string) (*IgReel, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.3 Safari/605.1.15")
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	r := regexp.MustCompile(`window.__additionalDataLoaded\('[\w\/]+',(.*?)\);</script>`)
	match := r.FindSubmatch(body)
	if len(match) < 1 {
		api.l.Error(match)
		return nil, fmt.Errorf("unexpected html")
	}

	var reel IgReel
	err = json.Unmarshal([]byte(match[1]), &reel)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	return &reel, nil
}

type IgReel struct {
	Items []IgReelItem
}

type IgReelUser struct {
	Username string
	FullName string `json:"full_name"`
}

type IgReelItem struct {
	Code          string
	User          IgReelUser
	Caption       IgReelCaption
	VideoVersions []IgReelVideo `json:"video_versions"`
}

type IgReelVideo struct {
	Width  int
	Height int
	URL    string
}

type IgReelCaption struct {
	Text string
}