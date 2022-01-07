package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokJsonAPI(l core.ILogger, hc core.IHttpClient, r core.IRand) core.IMediaFactory {
	return &TikTokJsonAPI{l, hc, r}
}

type TikTokJsonAPI struct {
	l  core.ILogger
	hc core.IHttpClient
	r  core.IRand
}

func (api *TikTokJsonAPI) CreateMedia(url string) ([]*core.Media, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected url %s", url)
	}
	apiURL := "https://www.tiktok.com/node/share/video/" + parts[3] + "/" + parts[5]
	api.l.Infof("processing %s", apiURL)
	api.hc.SetHeader("Cookie", "tt_webid_v2=69"+api.randomDigits(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	jsonString, err := api.hc.GetContent(apiURL)
	if err != nil {
		return nil, err
	}

	os.WriteFile("tiktok-"+parts[5]+".json", []byte(jsonString), 0644)

	var resp TikTokJSONResponse
	err = json.Unmarshal([]byte(jsonString), &resp)
	if err != nil {
		return nil, err
	}

	// if resp.Props.PageProps.ServerCode == 404 {
	// 	return nil, errors.New("Video currently unavailable")
	// }

	// if resp.Props.PageProps.StatusCode != 0 {
	// 	return nil, fmt.Errorf("%d not equal to zero", resp.Props.PageProps.StatusCode)
	// }

	item := resp.ItemInfo.ItemStruct
	title := item.Desc
	if len(title) == 0 {
		title = fmt.Sprintf("%s (@%s)", item.Author.Nickname, item.Author.UniqueId)
	}

	description := fmt.Sprintf("%s (@%s) has created a short video on TikTok with music %s.", item.Author.Nickname, item.Author.UniqueId, item.Music.Title)
	for _, s := range item.StickersOnItem {
		for _, t := range s.StickerText {
			description = description + " | " + t
		}
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: item.Video.DownloadAddr,
		Title:       title,
		Description: description,
	}

	return []*core.Media{media}, nil
}

func (api *TikTokJsonAPI) randomDigits(count int) string {
	rv := ""
	for i := 1; i < count; i++ {
		rv = rv + strconv.Itoa(api.r.GetRand(10))
	}
	return rv
}
