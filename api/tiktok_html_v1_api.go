package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokHTMLV1API(l core.ILogger, hc core.IHttpClient, r core.IRand) ITikTokAPI {
	return &TikTokHTMLV1API{l, hc, r}
}

type TikTokHTMLV1API struct {
	l  core.ILogger
	hc core.IHttpClient
	r  core.IRand
}

func (api *TikTokHTMLV1API) GetItem(username string, videoId string) (*TikTokItem, error) {
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

	var resp TikTokV1HTMLResponse
	err = json.Unmarshal([]byte(match[1]), &resp)
	if err != nil {
		return nil, err
	}

	if resp.Props.PageProps.ServerCode == 404 {
		return nil, fmt.Errorf("Video currently unavailable")
	}

	if resp.Props.PageProps.StatusCode != 0 {
		return nil, fmt.Errorf("%d not equal to zero", resp.Props.PageProps.StatusCode)
	}

	item := resp.Props.PageProps.ItemInfo.ItemStruct

	stickers := []string{}
	for _, s := range item.StickersOnItem {
		for _, t := range s.StickerText {
			stickers = append(stickers, t)
		}
	}

	i := TikTokItem{
		Author: TikTokAuthor{
			Nickname: item.Author.Nickname,
			UniqueId: item.Author.UniqueId,
		},
		Desc: item.Desc,
		Music: TikTokMusic{
			Title: item.Music.Title,
		},
		Stickers: stickers,
		VideoURL: item.Video.DownloadAddr,
	}

	return &i, nil
}

func (api *TikTokHTMLV1API) randomDigits(count int) string {
	rv := ""
	for i := 1; i < count; i++ {
		rv = rv + strconv.Itoa(api.r.GetRand(10))
	}
	return rv
}
