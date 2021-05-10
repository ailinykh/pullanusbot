package api

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	tb "gopkg.in/tucnak/telebot.v2"
)

type TelebotAdapter struct {
	m *tb.Message
	t *Telebot
}

func (a *TelebotAdapter) SendVideo(vf *core.VideoFile, caption string) error {
	video := makeVideoFile(vf, caption)
	a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
	_, err := video.Send(a.t.bot, a.m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		return err
	} else {
		a.t.logger.Infof("%s sent successfully", vf.FileName)
		a.t.bot.Delete(a.m)
	}
	return nil
}

func (a *TelebotAdapter) SendText(text string) error {
	_, err := a.t.bot.Send(a.m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	return err
}
