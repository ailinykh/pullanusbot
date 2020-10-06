package report

import (
	"os"
	"path"
	"pullanusbot/config"
	i "pullanusbot/interfaces"
	u "pullanusbot/utils"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

// Report is a helper to get logs simple way
type Report struct {
}

// Setup all nesessary command handlers
func (r *Report) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	bot.Handle("/logs", r.logs)
}

func (r *Report) logs(m *tb.Message) {
	conf := config.Get()

	if m.Chat.ID != conf.ReportChatID {
		logger.Warningf("ReportChatID mismatch! %d", m.Chat.ID)
		return
	}

	zipfile := os.TempDir() + "logs-" + u.RandStringRunes(4) + ".zip"
	defer os.Remove(zipfile)

	logfile := path.Join(conf.WorkingDir, "log.txt")

	err := ZipFiles(zipfile, []string{logfile})

	if err != nil {
		logger.Error(err)
		return
	}

	doc := tb.Document{File: tb.FromDisk(zipfile), FileName: "logs.zip"}
	doc.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{})
}
