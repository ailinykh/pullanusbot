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
	chanID int64
	chatID int64
	userID int

	photoChan chan *tb.Message
	actual    []tb.Message
}

// Setup all nesessary command handlers
func (p *Publisher) Setup(b i.Bot, conn *gorm.DB) {
	chanID, err := strconv.ParseInt(os.Getenv("PUBLISER_CHANNEL_ID"), 10, 64)
	if err != nil {
		logger.Error("PUBLISER_CHANNEL_ID not set")
		return
	}

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
	p.chanID = chanID
	p.chatID = chatID
	p.userID = userID
	p.photoChan = make(chan *tb.Message)
	bot = b
	bot.Handle("/loh666", p.loh666)

	go p.startLoop()

	logger.Info("successfully initialized")
}

// HandlePhoto is an i.PhotoHandler interface implementation
func (p *Publisher) HandlePhoto(m *tb.Message) {
	if m.Chat.ID != p.chatID || m.Sender.ID != p.userID {
		return
	}
	p.photoChan <- m
}

func (p *Publisher) loh666(m *tb.Message) {
	if m.Chat.ID != p.chatID {
		return
	}

	const timeount = 30 // seconds before message dissapears
	switch count := len(p.actual); count {
	case 0:
		bot.Send(m.Chat, "I have nothing for you comrade major")
	case 1:
		logger.Info("have one actual photo")
		sent, err := bot.Send(m.Chat, p.actual[0].Photo)
		if err != nil {
			logger.Error(err)
		} else {
			time.Sleep(time.Duration(timeount) * time.Second)
			bot.Delete(sent)
		}
	default:
		logger.Infof("have %d actual photos", count)
		album := tb.Album{}
		for _, m := range p.actual {
			album = append(album, m.Photo)
		}
		sent, err := bot.SendAlbum(m.Chat, album)
		if err != nil {
			logger.Error(err)
		} else {
			time.Sleep(time.Duration(timeount) * time.Second)
			for _, m := range sent {
				bot.Delete(&m)
			}
		}
	}
}

func (p *Publisher) startLoop() {
	queue := map[string][]*tb.Message{}

	for {
		select {
		case m := <-p.photoChan:
			if m.AlbumID == "" {
				sent, err := m.Photo.Send(bot.(*tb.Bot), &tb.Chat{ID: p.chanID}, &tb.SendOptions{})
				if err != nil {
					logger.Error(err)
				} else {
					logger.Info("photo published")
					p.actual = []tb.Message{*sent}
					bot.Delete(m)
				}
			} else {
				queue[m.AlbumID] = append(queue[m.AlbumID], m)
			}

		case <-time.After(1 * time.Second):
			for _, messages := range queue {
				album := tb.Album{}
				// No reason to sort it cause the Unixtime is
				// always the same for every Photo in Album
				// sort.Slice(messages, func(i, j int) bool {
				// 	return messages[i].Unixtime > messages[j].Unixtime
				// })
				for _, m := range messages {
					album = append(album, m.Photo)
				}
				var err error
				p.actual, err = bot.SendAlbum(&tb.Chat{ID: p.chanID}, album)
				if err != nil {
					logger.Error(err)
				} else {
					logger.Info("album published")
					for _, m := range messages {
						bot.Delete(m)
					}
				}
			}
			queue = map[string][]*tb.Message{}
		}
	}
}
