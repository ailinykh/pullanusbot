package api

import (
	"encoding/json"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokJsonAPI(l core.ILogger, hc core.IHttpClient, r core.IRand) ITikTokAPI {
	return &TikTokJsonAPI{l, hc, r}
}

type TikTokJsonAPI struct {
	l  core.ILogger
	hc core.IHttpClient
	r  core.IRand
}

func (api *TikTokJsonAPI) GetItem(username string, videoId string) (*TikTokV1ItemStruct, error) {
	url := "https://www.tiktok.com/node/share/video/" + username + "/" + videoId
	api.l.Infof("processing %s", url)
	api.hc.SetHeader("Cookie", "tt_webid_v2=69"+api.randomDigits(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	jsonString, err := api.hc.GetContent(url)
	if err != nil {
		return nil, err
	}

	var resp TikTokV1JSONResponse
	err = json.Unmarshal([]byte(jsonString), &resp)
	if err != nil {
		return nil, err
	}

	return &resp.ItemInfo.ItemStruct, nil
}

func (api *TikTokJsonAPI) randomDigits(count int) string {
	rv := ""
	for i := 1; i < count; i++ {
		rv = rv + strconv.Itoa(api.r.GetRand(10))
	}
	return rv
}
