package usecases

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokFlow(l core.ILogger, hc core.IHttpClient, sms core.ISendMediaStrategy) *TikTokFlow {
	hc.SetHeader("Referrer", "https://www.tiktok.com/")
	return &TikTokFlow{l, hc, sms}
}

type TikTokFlow struct {
	l   core.ILogger
	hc  core.IHttpClient
	sms core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (ttf *TikTokFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https?://\w+\.tiktok.com/\S+`)
	links := r.FindAllString(message.Text, -1)
	for _, l := range links {
		err := ttf.handleURL(l, message, bot)
		if err != nil {
			return err
		}
	}

	if len(links) > 0 {
		return bot.Delete(message)
	}
	return nil
}

func (ttf *TikTokFlow) handleURL(url string, message *core.Message, bot core.IBot) error {
	ttf.l.Info("processing ", url)
	fullURL, err := ttf.hc.GetRedirectLocation(url)
	if err != nil {
		return err
	}

	r := regexp.MustCompile(`tiktok\.com/(@\S+)/video/(\d+)`)
	match := r.FindStringSubmatch(fullURL)
	if len(match) != 3 {
		ttf.l.Error(match)
		return fmt.Errorf("unexpected redirect location %s", fullURL)
	}

	// apiURL := "https://www.tiktok.com/node/share/video/" + match[1] + "/" + match[2]
	apiURL := "https://www.tiktok.com/" + match[1] + "/video/" + match[2]
	ttf.l.Info(apiURL)

	getRand := func(count int) string {
		rv := ""
		for i := 1; i < count; i++ {
			rv = rv + strconv.Itoa(rand.Intn(10))
		}
		return rv
	}
	ttf.hc.SetHeader("Cookie", "tt_webid_v2=69"+getRand(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	htmlString, err := ttf.hc.GetContent(apiURL)
	if err != nil {
		return err
	}

	r = regexp.MustCompile(`<script id="__NEXT_DATA__" type="application\/json" nonce="[\w-]+" crossorigin="anonymous">(.*?)<\/script>`)
	match = r.FindStringSubmatch(htmlString)
	if len(match) < 1 {
		ttf.l.Error(match)
		return fmt.Errorf("unexpected html")
	}

	var resp TikTokHTMLResponse
	err = json.Unmarshal([]byte(match[1]), &resp)
	if err != nil {
		return err
	}

	if resp.Props.PageProps.StatusCode != 0 {
		return fmt.Errorf("%d not equal to zero", resp.Props.PageProps.StatusCode)
	}

	itemInfo := resp.Props.PageProps.ItemInfo
	title := itemInfo.ItemStruct.Desc
	if len(title) == 0 {
		title = fmt.Sprintf("%s (@%s)", itemInfo.ItemStruct.Author.Nickname, itemInfo.ItemStruct.Author.UniqueId)
	}

	description := fmt.Sprintf("%s (@%s) has created a short video on TikTok with music %s.", itemInfo.ItemStruct.Author.Nickname, itemInfo.ItemStruct.Author.UniqueId, itemInfo.ItemStruct.Music.Title)
	for _, s := range itemInfo.ItemStruct.StickersOnItem {
		for _, t := range s.StickerText {
			description = description + " | " + t
		}
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: itemInfo.ItemStruct.Video.DownloadAddr,
		Title:       title,
		Description: description,
	}
	media.Caption = fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", url, title, message.Sender.Username, description)
	return ttf.sms.SendMedia([]*core.Media{media}, bot)
}
