package api

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/helpers"
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
	videoHandlers    []core.IVideoHandler
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

	telebot := &Telebot{bot, logger, []string{}, []core.ITextHandler{}, []core.IDocumentHandler{}, []core.IImageHandler{}, []core.IVideoHandler{}}

	bot.Handle(tb.OnText, func(m *tb.Message) {
		for _, h := range telebot.textHandlers {
			err := h.HandleText(makeMessage(m), makeIBot(m, telebot))
			if err != nil && err.Error() != "not implemented" {
				logger.Errorf("%T: %s", h, err)
				telebot.reportError(m, err)
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

			logger.Infof("Downloaded to %s", strings.ReplaceAll(path, os.TempDir(), "$TMPDIR/"))
			defer os.Remove(path)

			for _, h := range telebot.documentHandlers {
				err := h.HandleDocument(&core.Document{
					File: core.File{Name: m.Document.FileName, Path: path},
					MIME: m.Document.MIME,
				}, makeMessage(m), makeIBot(m, telebot))
				if err != nil {
					logger.Errorf("%T: %s", h, err)
					telebot.reportError(m, err)
				}
			}
		}
	})

	bot.Handle(tb.OnPhoto, func(m *tb.Message) {

		image := &core.Image{
			ID:      m.Photo.FileID,
			FileURL: m.Photo.FileURL,
			Width:   m.Photo.Width,
			Height:  m.Photo.Height,
		}

		for _, h := range telebot.imageHandlers {
			err := h.HandleImage(image, makeMessage(m), makeIBot(m, telebot))
			if err != nil {
				logger.Errorf("%T: %s", h, err)
				telebot.reportError(m, err)
			}
		}
	})

	bot.Handle(tb.OnVideo, func(m *tb.Message) {

		video := &core.Video{
			ID:     m.Video.FileID,
			Width:  m.Video.Width,
			Height: m.Video.Height,
		}

		for _, h := range telebot.videoHandlers {
			err := h.HandleImage(video, makeMessage(m), makeIBot(m, telebot))
			if err != nil {
				logger.Errorf("%T: %s", h, err)
				telebot.reportError(m, err)
			}
		}
	})

	return telebot
}

// Download is a core.IImageDownloader interface implementation
func (t *Telebot) Download(image *core.Image) (*core.File, error) {
	//TODO: potential race condition
	file := tb.FromURL(image.FileURL)
	file.FileID = image.ID
	name := helpers.RandStringRunes(4) + ".jpg"
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
		if f, ok := handler[1].(func(*core.Message, core.IBot) error); ok {
			t.bot.Handle(h, func(m *tb.Message) {
				f(makeMessage(m), &TelebotAdapter{m, t})
			})
		} else {
			panic("interface must implement func(*core.Message, core.IBot) error")
		}
	default:
		panic(fmt.Sprintf("something wrong with %s", h))
	}

	if h, ok := handler[0].(core.IButtonHandler); ok {
		for _, id := range h.GetButtonIds() {
			t.bot.Handle("\f"+id, func(c *tb.Callback) {
				err := h.ButtonPressed(c.Data, makeMessage(c.Message), makeIBot(c.Message, t))
				if err != nil {
					t.logger.Error(err)
					t.reportError(c.Message, err)
				}
				t.bot.Respond(c, &tb.CallbackResponse{CallbackID: c.ID})
			})
		}
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

func (t *Telebot) reportError(m *tb.Message, e error) {
	chatID, err := strconv.ParseInt(os.Getenv("ADMIN_CHAT_ID"), 10, 64)
	if err != nil {
		return
	}
	chat := &tb.Chat{ID: chatID}
	opts := &tb.SendOptions{DisableWebPagePreview: true}
	t.bot.Forward(chat, m, opts)
	t.bot.Send(chat, e.Error(), opts)
}

func makeMessage(m *tb.Message) *core.Message {
	text := m.Text
	if m.Document != nil {
		text = m.Caption
	}
	message := &core.Message{
		ID:        m.ID,
		ChatID:    m.Chat.ID,
		IsPrivate: m.Private(),
		Sender:    makeUser(m.Sender),
		Text:      text,
	}

	if m.ReplyTo != nil {
		message.ReplyTo = makeMessage(m.ReplyTo)
	}

	if m.Video != nil {
		message.Video = makeVideo(m.Video)
	}

	return message
}

func makeUser(u *tb.User) *core.User {
	return &core.User{
		ID:           u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     u.Username,
		LanguageCode: u.LanguageCode,
	}
}

func makeVideo(v *tb.Video) *core.Video {
	return &core.Video{
		File: core.File{
			Name: v.FileName,
			Path: v.FileURL,
		},
		ID:       v.FileID,
		Width:    v.Width,
		Height:   v.Height,
		Bitrate:  0,
		Duration: v.Duration,
		Codec:    "",
		Thumb:    makePhoto(v.Thumbnail),
	}
}

func makePhoto(p *tb.Photo) *core.Image {
	return &core.Image{
		File: core.File{
			Name: p.FileLocal,
			Path: p.FilePath,
		},
		ID:      p.FileID,
		FileURL: p.FileURL,
		Width:   p.Width,
		Height:  p.Height,
	}
}

func makeFile(name string, path string) *core.File {
	return &core.File{
		Name: name,
		Path: path,
	}
}

func makeIBot(m *tb.Message, t *Telebot) core.IBot {
	return &TelebotAdapter{m, t}
}
