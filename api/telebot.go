package api

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Telebot struct {
	bot              *tb.Bot
	logger           core.ILogger
	commandHandlers  []string
	textHandlers     []core.ITextHandler
	documentHandlers []core.IDocumentHandler
	imageHandlers    []core.IImageHandler
}

func CreateTelebot(token string, logger core.ILogger) *Telebot {
	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	var err error
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		panic(err)
	}

	telebot := &Telebot{bot, logger, []string{}, []core.ITextHandler{}, []core.IDocumentHandler{}, []core.IImageHandler{}}

	bot.Handle(tb.OnText, func(m *tb.Message) {
		for _, h := range telebot.textHandlers {
			err := h.HandleText(m.Text, makeUser(m), &TelebotAdapter{m, telebot})
			if err != nil {
				logger.Error(err)
			}
		}
	})

	var mutex sync.Mutex

	bot.Handle(tb.OnDocument, func(m *tb.Message) {
		// TODO: inject `download` to get rid of MIME cheking
		if m.Document.MIME[:5] == "video" || m.Document.MIME == "image/gif" {
			mutex.Lock()
			defer mutex.Unlock()

			logger.Infof("Attempt to download %s %s (sent by %s)", m.Document.FileName, m.Document.MIME, m.Sender.Username)

			path := path.Join(os.TempDir(), m.Document.FileName)
			err := bot.Download(&m.Document.File, path)
			if err != nil {
				logger.Error(err)
				return
			}

			logger.Infof("Downloaded to %s", path)
			defer os.Remove(path)

			for _, h := range telebot.documentHandlers {
				h.HandleDocument(&core.Document{
					Author:   m.Sender.Username,
					FileName: m.Document.FileName,
					FilePath: path,
					MIME:     m.Document.MIME,
				}, &TelebotAdapter{m, telebot})
			}
		}
	})

	bot.Handle(tb.OnPhoto, func(m *tb.Message) {

		image := core.CreateImage(m.Photo.FileID, func() (string, error) {
			path := path.Join(os.TempDir(), m.Photo.File.UniqueID+".jpg")
			err := bot.Download(&m.Photo.File, path)
			if err != nil {
				logger.Error(err)
				return "", err
			}

			logger.Infof("image %s downloaded to %s", m.Photo.UniqueID, path)
			return path, nil
		})
		defer image.Dispose()

		for _, h := range telebot.imageHandlers {
			h.HandleImage(&image, createMessage(m), &TelebotAdapter{m, telebot})
		}
	})

	return telebot
}

func (t *Telebot) AddHandler(handlers ...interface{}) {
	switch h := handlers[0].(type) {
	case core.IDocumentHandler:
		t.documentHandlers = append(t.documentHandlers, h)
	case core.ITextHandler:
		t.textHandlers = append(t.textHandlers, h)
	case core.IImageHandler:
		t.imageHandlers = append(t.imageHandlers, h)
	case string:
		t.registerCommand(h)
		if handler, ok := handlers[1].(core.ICommandHandler); ok {
			t.bot.Handle(h, func(m *tb.Message) {
				handler.HandleCommand(createMessage(m), &TelebotAdapter{m, t})
			})
		} else {
			panic("interface must implement core.ICommandHandler")
		}
	default:
		panic(fmt.Sprintf("something wrong with %s", h))
	}
}

func (t *Telebot) Run() {
	t.bot.Start()
}

func (t *Telebot) registerCommand(command string) {
	for _, c := range t.commandHandlers {
		if c == command {
			panic("Handler for " + command + " already set!")
		}
	}
	t.commandHandlers = append(t.commandHandlers, command)
}

func createMessage(m *tb.Message) *core.Message {
	return &core.Message{
		ID:        m.ID,
		ChatID:    m.Chat.ID,
		IsPrivate: m.Private(),
		Sender:    &core.User{Username: m.Sender.Username},
		Text:      m.Text,
	}
}
