package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokAPI(l core.ILogger, hc core.IHttpClient, r core.IRand) *TikTokAPI {
	return &TikTokAPI{l, hc, r}
}

type TikTokAPI struct {
	l  core.ILogger
	hc core.IHttpClient
	r  core.IRand
}

func (api *TikTokAPI) getItemUsingJsonApi(username string, videoId string) (*TikTokItemStruct, error) {
	url := "https://www.tiktok.com/node/share/video/" + username + "/" + videoId
	api.l.Infof("processing %s", url)
	api.hc.SetHeader("Cookie", "tt_webid_v2=69"+api.randomDigits(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	jsonString, err := api.hc.GetContent(url)
	if err != nil {
		return nil, err
	}

	var resp TikTokJSONResponse
	err = json.Unmarshal([]byte(jsonString), &resp)
	if err != nil {
		return nil, err
	}

	return &resp.ItemInfo.ItemStruct, nil
}

func (api *TikTokAPI) getItemUsingHtmlApi(username string, videoId string) (*TikTokItemStruct, error) {
	url := "https://www.tiktok.com/" + username + "/video/" + videoId
	api.l.Infof("processing %s", url)
	api.hc.SetHeader("Cookie", "tt_webid_v2=69"+api.randomDigits(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	htmlString, err := api.hc.GetContent(url)
	if err != nil {
		return nil, err
	}

	// os.WriteFile("tiktok-"+strings.Split(url, "/")[5]+".html", []byte(htmlString), 0644)
	r := regexp.MustCompile(`<script id="__NEXT_DATA__" type="application\/json" nonce="[\w-]+" crossorigin="anonymous">(.*?)<\/script>`)
	match := r.FindStringSubmatch(htmlString)
	if len(match) < 1 {
		api.l.Error(match)
		return nil, fmt.Errorf("unexpected html")
	}

	var resp TikTokHTMLResponse
	err = json.Unmarshal([]byte(match[1]), &resp)
	if err != nil {
		return nil, err
	}

	if resp.Props.PageProps.ServerCode == 404 {
		return nil, errors.New("Video currently unavailable")
	}

	if resp.Props.PageProps.StatusCode != 0 {
		return nil, fmt.Errorf("%d not equal to zero", resp.Props.PageProps.StatusCode)
	}

	return &resp.Props.PageProps.ItemInfo.ItemStruct, nil
}

func (api *TikTokAPI) randomDigits(count int) string {
	rv := ""
	for i := 1; i < count; i++ {
		rv = rv + strconv.Itoa(api.r.GetRand(10))
	}
	return rv
}
