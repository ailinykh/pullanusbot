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

func (a *TelebotAdapter) SendText(text string, params ...interface{}) (*core.Message, error) {
	opts := tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true}
	for _, param := range params {
		switch m := param.(type) {
		case *core.Message:
			opts.ReplyTo = &tb.Message{ID: m.ID}
		default:
			break
		}
	}
	sent, err := a.t.bot.Send(a.m.Chat, text, &opts)
	return makeMessage(sent), err
}

func (a *TelebotAdapter) Delete(message *core.Message) error {
	return a.t.bot.Delete(&tb.Message{ID: message.ID, Chat: &tb.Chat{ID: message.ChatID}})
}

func (a *TelebotAdapter) SendImage(image *core.Image) (*core.Message, error) {
	photo := &tb.Photo{File: tb.File{FileID: image.ID}}
	sent, err := a.t.bot.Send(a.m.Chat, photo)
	return makeMessage(sent), err
}

func (a *TelebotAdapter) SendAlbum(images []*core.Image) ([]*core.Message, error) {
	album := tb.Album{}
	for _, i := range images {
		photo := &tb.Photo{File: tb.File{FileID: i.ID}}
		album = append(album, photo)
	}

	sent, err := a.t.bot.SendAlbum(a.m.Chat, album)
	var messages []*core.Message
	for _, m := range sent {
		messages = append(messages, makeMessage(&m))
	}
	return messages, err
}

func (a *TelebotAdapter) SendMedia(media *core.Media) (*core.Message, error) {
	var sent *tb.Message
	var err error
	switch media.Type {
	case core.Photo:
		file := &tb.Photo{File: tb.FromURL(media.URL)}
		file.Caption = media.Caption
		a.t.bot.Notify(a.m.Chat, tb.UploadingPhoto)
		sent, err = a.t.bot.Send(a.m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	case core.Video:
		file := &tb.Video{File: tb.FromURL(media.URL)}
		file.Caption = media.Caption
		a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
		sent, err = a.t.bot.Send(a.m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	case core.Text:
		sent, err = a.t.bot.Send(a.m.Chat, media.Caption, &tb.SendOptions{ParseMode: tb.ModeHTML})
	}

	return makeMessage(sent), err
}

func (a *TelebotAdapter) SendPhotoAlbum(medias []*core.Media) ([]*core.Message, error) {
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

	sent, err := a.t.bot.SendAlbum(a.m.Chat, album)
	var messages []*core.Message
	for _, m := range sent {
		messages = append(messages, makeMessage(&m))
	}
	return messages, err
}

func (a *TelebotAdapter) SendVideoFile(vf *core.VideoFile, caption string) (*core.Message, error) {
	video := makeVideoFile(vf, caption)
	a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
	sent, err := video.Send(a.t.bot, a.m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err == nil {
		a.t.logger.Infof("%s sent successfully", vf.Name)
		a.t.bot.Delete(a.m)
	}
	return makeMessage(sent), err
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
