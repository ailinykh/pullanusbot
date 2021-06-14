package use_cases

import (
	"os"
	"strconv"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreatePublisherFlow(l core.ILogger) *PublisherFlow {
	chatID, err := strconv.ParseInt(os.Getenv("PUBLISER_CHAT_ID"), 10, 64)
	if err != nil {
		chatID = 0
	}

	username := os.Getenv("PUBLISER_USERNAME")

	publisher := PublisherFlow{
		l:           l,
		chatID:      chatID,
		username:    username,
		imageChan:   make(chan imgSource),
		requestChan: make(chan msgSource),
	}

	go publisher.runLoop()
	return &publisher
}

type PublisherFlow struct {
	l core.ILogger

	chatID      int64
	username    string
	imageChan   chan imgSource
	requestChan chan msgSource
}

type imgSource struct {
	imageID string
	bot     core.IBot
}

type msgSource struct {
	message core.Message
	bot     core.IBot
}

// core.IImageHandler
func (p *PublisherFlow) HandleImage(image *core.Image, message *core.Message, bot core.IBot) error {
	if message.ChatID != p.chatID || message.Sender.Username != p.username {
		return nil
	}

	p.imageChan <- imgSource{image.ID, bot}
	return nil
}

// core.ICommandHandler
func (p *PublisherFlow) HandleCommand(message *core.Message, bot core.IBot) error {
	if message.ChatID != p.chatID {
		return nil
	}

	p.requestChan <- msgSource{*message, bot}
	return nil
}

func (p *PublisherFlow) runLoop() {
	photos := []string{}
	queue := []string{}

	disposal := func(m core.Message, bot core.IBot, timeout int) {
		time.Sleep(time.Duration(timeout) * time.Second)
		p.l.Infof("disposing message %d from chat %d", m.ID, m.ChatID)
		err := bot.Delete(&m)
		if err != nil {
			p.l.Error(err)
		}
	}

	for {
		select {
		case is := <-p.imageChan:
			p.l.Infof("got photo %s", is.imageID)
			queue = append(queue, is.imageID)

		case <-time.After(1 * time.Second):
			if len(queue) > 0 {
				p.l.Infof("had %d actual photo(s)", len(queue))
				photos = queue
				queue = []string{}
			}

		case ms := <-p.requestChan:
			go disposal(ms.message, ms.bot, 0)
			switch count := len(photos); count {
			case 0:
				err := ms.bot.SendText("I have nothing for you comrade major")
				if err != nil {
					p.l.Error(err)
				}
			case 1:
				p.l.Info("have one actual photo")
				sent, err := ms.bot.SendImage(&core.Image{ID: photos[0]})
				if err != nil {
					p.l.Error(err)
				} else {
					go disposal(*sent, ms.bot, 30)
				}
			default:
				p.l.Infof("have %d actual photos", count)
				album := []*core.Image{}
				for _, p := range photos {
					album = append(album, &core.Image{ID: p})
				}
				sent, err := ms.bot.SendAlbum(album)
				if err != nil {
					p.l.Error(err)
				} else {
					for _, m := range sent {
						go disposal(*m, ms.bot, 30)
					}
				}
			}
		}
	}
}
