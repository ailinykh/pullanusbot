package main

import (
	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// ITextHandler needed to get text messages
type ITextHandler interface {
	handleTextMessage(m *tb.Message)
}

// TextHandler is a common message handler
type TextHandler struct {
	handlers []ITextHandler
}

func (h *TextHandler) initialize() {
	bot.Handle(tb.OnText, h.handleText)
	logger.Info("successfully initialized")
}

func (h *TextHandler) handleText(m *tb.Message) {
	for _, handler := range h.handlers {
		handler.handleTextMessage(m)
	}
}
