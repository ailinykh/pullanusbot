package api

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	tb "gopkg.in/tucnak/telebot.v2"
)

func makeTbVideo(vf *core.Video, caption string) *tb.Video {
	var video *tb.Video
	if len(vf.ID) > 0 {
		video = &tb.Video{File: tb.File{FileID: vf.ID}}
	} else {
		video = &tb.Video{File: tb.FromDisk(vf.Path)}
		video.FileName = vf.File.Name
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
	}
	return video
}

func makeTbPhoto(image *core.Image, caption string) *tb.Photo {
	photo := &tb.Photo{File: tb.FromDisk(image.File.Path)}
	if len(image.ID) > 0 {
		photo = &tb.Photo{File: tb.File{FileID: image.ID}}
	}
	photo.Caption = caption
	return photo
}
