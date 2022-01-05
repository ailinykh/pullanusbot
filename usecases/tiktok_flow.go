package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokFlow(l core.ILogger, hc core.IHttpClient, sms core.ISendMediaStrategy, api ITikTokAPI) *TikTokFlow {
	hc.SetHeader("Referrer", "https://www.tiktok.com/")
	return &TikTokFlow{l, hc, sms, api}
}

type ITikTokAPI interface {
	Get(string) (*TikTokResponse, error)
}

type TikTokFlow struct {
	l   core.ILogger
	hc  core.IHttpClient
	sms core.ISendMediaStrategy
	api ITikTokAPI
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

	resp, err := ttf.api.Get(originalURL)
	if err != nil {
		return err
	}

	if resp.ServerCode == 404 {
		_, err := bot.SendText(originalURL + "\nVideo currently unavailable")
		return err
	}

	if resp.StatusCode != 0 {
		ttf.l.Error(match[1])
		return fmt.Errorf("%d not equal to zero", resp.StatusCode)
	}

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
	media.Caption = fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", url, title, message.Sender.DisplayName(), description)
	return ttf.sms.SendMedia([]*core.Media{media}, bot)
}
