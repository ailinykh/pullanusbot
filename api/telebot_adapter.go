package api

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	tb "gopkg.in/tucnak/telebot.v2"
)

type TelebotAdapter struct {
	m *tb.Message
	t *Telebot
}

func (a *TelebotAdapter) SendText(text string) error {
	_, err := a.t.bot.Send(a.m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true})
	return err
}

func (a *TelebotAdapter) SendPhoto(media *core.Media) error {
	file := &tb.Photo{File: tb.FromURL(media.URL)}
	file.Caption = media.Caption
	a.t.bot.Notify(a.m.Chat, tb.UploadingPhoto)
	_, err := a.t.bot.Send(a.m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	return err
}

func (a *TelebotAdapter) SendPhotoAlbum(medias []*core.Media) error {
	var photo *tb.Photo
	var album = tb.Album{}

	for i, m := range medias {
		photo = &tb.Photo{File: tb.FromURL(m.URL)}
		if i == len(medias)-1 {
			photo.Caption = m.Caption
			photo.ParseMode = tb.ModeHTML
		}
		album = append(album, photo)
	}

	_, err := a.t.bot.SendAlbum(a.m.Chat, album)
	return err
}

func (a *TelebotAdapter) SendVideo(media *core.Media) error {
	file := &tb.Video{File: tb.FromURL(media.URL)}
	file.Caption = media.Caption
	a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
	_, err := a.t.bot.Send(a.m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	return err
}

func (a *TelebotAdapter) SendVideoFile(vf *core.VideoFile, caption string) error {
	video := makeVideoFile(vf, caption)
	a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
	_, err := video.Send(a.t.bot, a.m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		return err
	} else {
		a.t.logger.Infof("%s sent successfully", vf.Name)
		a.t.bot.Delete(a.m)
	}
	return nil
}

func (a *TelebotAdapter) CreatePlayer(string) infrastructure.Player {
	return infrastructure.Player{
		GameID:       a.m.Chat.ID,
		UserID:       a.m.Sender.ID,
		FirstName:    a.m.Sender.FirstName,
		LastName:     a.m.Sender.LastName,
		Username:     a.m.Sender.Username,
		LanguageCode: a.m.Sender.LanguageCode,
	}
}
