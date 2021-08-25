package usecases

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokFlow(l core.ILogger, hc core.IHttpClient, mf core.IMediaFactory, sms core.ISendMediaStrategy) *TikTokFlow {
	return &TikTokFlow{l, hc, mf, sms}
}

type TikTokFlow struct {
	l   core.ILogger
	hc  core.IHttpClient
	mf  core.IMediaFactory
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

	apiURL := "https://www.tiktok.com/node/share/video/" + match[1] + "/" + match[2]
	jsonString, err := ttf.hc.GetContent(apiURL)
	if err != nil {
		return err
	}

	var resp TikTokResponse
	err = json.Unmarshal([]byte(jsonString), &resp)
	if err != nil {
		return err
	}

	if resp.StatusCode != 0 {
		ttf.l.Error(jsonString)
		return fmt.Errorf("%d not equal to zero", resp.StatusCode)
	}

	title := resp.ItemInfo.ItemStruct.Desc
	if len(title) == 0 {
		title = fmt.Sprintf("%s (@%s)", resp.ItemInfo.ItemStruct.Author.Nickname, resp.ItemInfo.ItemStruct.Author.UniqueId)
	}

	description := fmt.Sprintf("%s (@%s) has created a short video on TikTok with music %s.", resp.ItemInfo.ItemStruct.Author.Nickname, resp.ItemInfo.ItemStruct.Author.UniqueId, resp.ItemInfo.ItemStruct.Music.Title)
	for _, s := range resp.ItemInfo.ItemStruct.StickersOnItem {
		for _, t := range s.StickerText {
			description = description + " | " + t
		}
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: resp.ItemInfo.ItemStruct.Video.DownloadAddr,
		Title:       title,
		Description: description,
	}
	media.Caption = fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", url, title, message.Sender.Username, description)
	return ttf.sms.SendMedia([]*core.Media{media}, bot)
}
