package main

import (
	"log"
	"math/rand"
	"os"
	"path"
	"time"

	"pullanusbot/config"
	"pullanusbot/converter"
	"pullanusbot/faggot"
	"pullanusbot/info"
	i "pullanusbot/interfaces"
	"pullanusbot/link"
	"pullanusbot/report"
	"pullanusbot/smsreg"
	"pullanusbot/telegraph"
	"pullanusbot/twitter"
	"pullanusbot/youtube"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
)

var (
	bot i.Bot
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	conf := config.Get()

	if _, err := os.Stat(conf.WorkingDir); os.IsNotExist(err) {
		os.MkdirAll(conf.WorkingDir, os.ModePerm)
	}

	logPath := path.Join(conf.WorkingDir, "log.txt")
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	defer lf.Close()
	defer logger.Init("pullanusbot", true, false, lf).Close()

	dbFile := path.Join(conf.WorkingDir, "pullanusbot.db")
	logger.Info("Using database: ", dbFile)
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: loger.Default.LogMode(loger.Error),
	})
	if err != nil {
		log.Fatal(err)
	}

	setupBot(conf.BotToken)

	adapters := []i.IBotAdapter{
		&faggot.Game{},
		&converter.Converter{},
		&info.Info{},
		&link.Link{},
		&report.Report{},
		&smsreg.SmsReg{},
		&telegraph.Telegraph{},
		&twitter.Twitter{},
		&youtube.Youtube{},
	}

	for _, adapter := range adapters {
		adapter.Setup(bot, conn)
	}

	bot.Handle(tb.OnText, func(m *tb.Message) {
		for _, adapter := range adapters {
			if handler, ok := adapter.(i.TextMessageHandler); ok {
				handler.HandleTextMessage(m)
			}
		}
	})

	bot.Start()
}

func setupBot(token string) {
	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	var err error
	bot, err = tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
		URL:    os.Getenv("API_URL"),
	})

	if err != nil {
		panic(err)
	}
}
