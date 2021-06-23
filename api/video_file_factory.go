package api

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	tb "gopkg.in/tucnak/telebot.v2"
)

func makeVideo(vf *core.Video, caption string) tb.Video {
	video := tb.Video{File: tb.FromDisk(vf.Path)}
	video.Width = vf.Width
	video.Height = vf.Height
	video.Caption = caption
	video.Duration = vf.Duration
	video.SupportsStreaming = true
	video.Thumbnail = &tb.Photo{
		File:   tb.FromDisk(vf.Thumb.Path),
		Width:  vf.Thumb.Width,
		Height: vf.Thumb.Height,
	}
	return video
}
