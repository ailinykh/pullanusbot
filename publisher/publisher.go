package publisher

import (
	"os"
	i "pullanusbot/interfaces"
	"strconv"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

// Publisher is ...
type Publisher struct {
	chatID int64
	userID int

	photoChan    chan tb.Message
	requestChan  chan tb.Message
	disposalChan chan tb.Message
}

// Setup all nesessary command handlers
func (p *Publisher) Setup(b i.Bot, conn *gorm.DB) {
	chatID, err := strconv.ParseInt(os.Getenv("PUBLISER_CHAT_ID"), 10, 64)
	if err != nil {
		logger.Error("PUBLISER_CHAT_ID not set")
		return
	}

	userID, err := strconv.Atoi(os.Getenv("PUBLISER_USER_ID"))
	if err != nil {
		logger.Error("PUBLISER_USER_ID not set")
		return
	}
	p.chatID = chatID
	p.userID = userID
	p.photoChan = make(chan tb.Message)
	p.requestChan = make(chan tb.Message)
	p.disposalChan = make(chan tb.Message)
	bot = b
	bot.Handle("/loh666", p.loh666)

	go p.startListener()
	go p.startDisposal()

	logger.Info("successfully initialized")
}

// HandlePhoto is an i.PhotoHandler interface implementation
func (p *Publisher) HandlePhoto(m *tb.Message) {
	if m.Chat.ID != p.chatID || m.Sender.ID != p.userID {
		return
	}

	p.photoChan <- *m
}

func (p *Publisher) loh666(m *tb.Message) {
	if m.Chat.ID != p.chatID {
		return
	}

	p.requestChan <- *m
}

func (p *Publisher) startListener() {
	photos := []string{}
	queue := []string{}

	for {
		select {
		case m := <-p.photoChan:
			logger.Infof("got album %s photo %s", m.AlbumID, m.Photo.FileID)
			queue = append(queue, m.Photo.FileID)

		case <-time.After(1 * time.Second):
			if len(queue) > 0 {
				logger.Infof("had %d actual photo(s)", len(queue))
				photos = queue
				queue = []string{}
			}

		case m := <-p.requestChan:
			switch count := len(photos); count {
			case 0:
				bot.Send(m.Chat, "I have nothing for you comrade major")
			case 1:
				logger.Info("have one actual photo")
				photo := &tb.Photo{File: tb.File{FileID: photos[0]}}
				sent, err := bot.Send(m.Chat, photo)
				if err != nil {
					logger.Error(err)
				} else {
					p.disposalChan <- *sent
				}
			default:
				logger.Infof("have %d actual photos", count)
				album := tb.Album{}
				for _, p := range photos {
					photo := &tb.Photo{File: tb.File{FileID: p}}
					album = append(album, photo)
				}
				sent, err := bot.SendAlbum(m.Chat, album)
				if err != nil {
					logger.Error(err)
				} else {
					for _, m := range sent {
						p.disposalChan <- m
					}
				}
			}
		}
	}
}

func (p *Publisher) startDisposal() {
	const timeout = 30 // seconds before message dissapears
	for {
		select {
		case m := <-p.disposalChan:
			logger.Infof("disposing %d %d", m.Chat.ID, m.ID)
			go func(m tb.Message) {
				time.Sleep(timeout * time.Second)
				err := bot.Delete(&m)
				if err != nil {
					logger.Error(err)
				}
			}(m)
		}
	}
}
