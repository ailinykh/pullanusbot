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
	ttf.l.Infof("processing %s", url)
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
	originalURL := "https://www.tiktok.com/" + match[1] + "/video/" + match[2]
	ttf.l.Infof("original: %s", originalURL)

	resp, err := ttf.retreiveContentFrom(originalURL)
	if err != nil {
		return err
	}

	if resp.Props.PageProps.ServerCode == 404 {
		_, err := bot.SendText(originalURL + "\nVideo currently unavailable")
		return err
	}

	if resp.Props.PageProps.StatusCode != 0 {
		ttf.l.Error(match[1])
		return fmt.Errorf("%d not equal to zero", resp.Props.PageProps.StatusCode)
	}

	item := resp.Props.PageProps.ItemInfo.ItemStruct
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
	media.Caption = fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", url, title, message.Sender.Username, description)
	return ttf.sms.SendMedia([]*core.Media{media}, bot)
}

func (ttf *TikTokFlow) retreiveContentFrom(url string) (*TikTokHTMLResponse, error) {
	var resp *TikTokHTMLResponse
	getRand := func(count int) string {
		rv := ""
		for i := 1; i < count; i++ {
			rv = rv + strconv.Itoa(rand.Intn(10))
		}
		return rv
	}
	ttf.hc.SetHeader("Cookie", "tt_webid_v2=69"+getRand(17)+"; Domain=tiktok.com; Path=/; Secure; hostOnly=false; hostOnly=false; aAge=4ms; cAge=4ms")
	htmlString, err := ttf.hc.GetContent(url)
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile(`<script id="__NEXT_DATA__" type="application\/json" nonce="[\w-]+" crossorigin="anonymous">(.*?)<\/script>`)
	match := r.FindStringSubmatch(htmlString)
	if len(match) < 1 {
		ttf.l.Error(match)
		return nil, fmt.Errorf("unexpected html")
	}

	err = json.Unmarshal([]byte(match[1]), &resp)
	if err != nil {
		return nil, err
	}

	return resp, err
}
