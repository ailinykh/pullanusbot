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

// Telebot is a telegram API
type Telebot struct {
	bot              *tb.Bot
	logger           core.ILogger
	commandHandlers  []string
	textHandlers     []core.ITextHandler
	documentHandlers []core.IDocumentHandler
	imageHandlers    []core.IImageHandler
}

// CreateTelebot is a default Telebot factory
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
			err := h.HandleText(makeMessage(m), &TelebotAdapter{m, telebot})
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

		image := core.CreateImage(m.Photo.FileID, m.Photo.FileURL)

		for _, h := range telebot.imageHandlers {
			h.HandleImage(&image, makeMessage(m), &TelebotAdapter{m, telebot})
		}
	})

	return telebot
}

// Download is a core.IImageDownloader interface implementation
func (t *Telebot) Download(image *core.Image) (*core.File, error) {
	//TODO: potential race condition
	file := tb.FromURL(image.FileURL)
	file.FileID = image.ID
	name := RandStringRunes(4) + ".jpg"
	path := path.Join(os.TempDir(), name)
	err := t.bot.Download(&file, path)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}

	t.logger.Infof("image %s downloaded to %s", file.UniqueID, path)
	return makeFile(name, path), nil
}

// AddHandler register object as one of core.Handler's
func (t *Telebot) AddHandler(handler ...interface{}) {
	switch h := handler[0].(type) {
	case core.IDocumentHandler:
		t.documentHandlers = append(t.documentHandlers, h)
	case core.ITextHandler:
		t.textHandlers = append(t.textHandlers, h)
	case core.IImageHandler:
		t.imageHandlers = append(t.imageHandlers, h)
	case string:
		t.registerCommand(h)
		if handler, ok := handler[1].(core.ICommandHandler); ok {
			t.bot.Handle(h, func(m *tb.Message) {
				handler.HandleCommand(makeMessage(m), &TelebotAdapter{m, t})
			})
		} else {
			panic("interface must implement core.ICommandHandler")
		}
	default:
		panic(fmt.Sprintf("something wrong with %s", h))
	}
}

// Run bot loop
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

func makeMessage(m *tb.Message) *core.Message {
	return &core.Message{
		ID:        m.ID,
		ChatID:    m.Chat.ID,
		IsPrivate: m.Private(),
		Sender:    &core.User{Username: m.Sender.Username},
		Text:      m.Text,
	}
}

func makeFile(name string, path string) *core.File {
	return &core.File{
		Name: name,
		Path: path,
	}
}
