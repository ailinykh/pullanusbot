package main

import (
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestConverterProceedOnlyVideoFiles(t *testing.T) {
	defer tearUp(t)()
	c := Converter{}
	c.initialize()
	m := &tb.Message{}
	m.Document = &tb.Document{MIME: "audio/mp3"}

	c.checkMessage(m)
}

func TestConverterProceedOnlyVideoFilesOver20MB(t *testing.T) {
	defer tearUp(t)()
	c := Converter{}
	c.initialize()
	m := &tb.Message{Sender: &tb.User{Username: "hitler"}}
	// file := tb.File{FileSize: 20971520}
	// file := tb.File{}
	m.Document = &tb.Document{MIME: "video/mp4", FileName: "video.mp4", File: tb.File{FileSize: 20971521}}

	c.checkMessage(m)
}

func TestConverterBotCastFailing(t *testing.T) {
	defer tearUp(t)()
	c := Converter{}
	c.initialize()
	m := &tb.Message{Sender: &tb.User{Username: "hitler"}}
	m.Document = &tb.Document{MIME: "video/mp4", FileName: "video.mp4", File: tb.File{FileSize: 20971520}}

	c.checkMessage(m)
}
