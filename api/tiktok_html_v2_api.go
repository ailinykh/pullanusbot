package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokHTMLV2API(l core.ILogger, hc core.IHttpClient, r core.IRand) ITikTokAPI {
	return &TikTokHTMLV2API{l, hc, r}
}

type TikTokHTMLV2API struct {
	l  core.ILogger
	hc core.IHttpClient
	r  core.IRand
}

func (api *TikTokHTMLV2API) GetItem(username string, videoId string) (*TikTokItem, error) {
	url := "https://www.tiktok.com/" + username + "/video/" + videoId
	api.l.Infof("processing %s", url)
	api.hc.SetHeader("Cookie", "tt_webid_v2=69"+api.randomDigits(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	htmlString, err := api.hc.GetContent(url)
	if err != nil {
		return nil, err
	}

	// os.WriteFile("tiktok-"+strings.Split(url, "/")[5]+".html", []byte(htmlString), 0644)
	r := regexp.MustCompile(`<script id="SIGI_STATE" type="application\/json">(.*?)<\/script>`)
	match := r.FindStringSubmatch(htmlString)
	if len(match) < 1 {
		api.l.Error(match)
		return nil, fmt.Errorf("unexpected html")
	}

	var resp TikTokV2HTMLNResponse
	err = json.Unmarshal([]byte(match[1]), &resp)
	if err != nil {
		return nil, err
	}

	if resp.VideoPage.StatusCode != 0 {
		return nil, fmt.Errorf("unextected status code %d", resp.VideoPage.StatusCode)
	}
	// os.WriteFile("tiktok-"+strings.Split(url, "/")[5]+".json", []byte(match[1]), 0644)
	item := resp.ItemModule[videoId]

	stickers := []string{}
	for _, s := range item.StickersOnItem {
		for _, t := range s.StickerText {
			stickers = append(stickers, t)
		}
	}

	author := TikTokAuthor{}
	if a, ok := resp.UserModule.Users[username[1:]]; ok {
		author.Nickname = a.Nickname
		author.UniqueId = a.UniqueId
	}

	i := TikTokItem{
		Author: author,
		Desc:   item.Desc,
		Music: TikTokMusic{
			Title: item.Music.Title,
		},
		Stickers: stickers,
		VideoURL: item.Video.DownloadAddr,
	}

	return &i, nil
}

func (api *TikTokHTMLV2API) randomDigits(count int) string {
	rv := ""
	for i := 1; i < count; i++ {
		rv = rv + strconv.Itoa(api.r.GetRand(10))
	}
	return rv
}
