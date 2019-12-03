package main

import (
	"database/sql"
	"log"
	"os"
	"path"
	"time"

	"github.com/google/logger"
	_ "github.com/mattn/go-sqlite3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// IBot is a generic interface for testing
type IBot interface {
	Handle(interface{}, interface{})
	Send(tb.Recipient, interface{}, ...interface{}) (*tb.Message, error)
	Start()
}

// IBotAdapter is a generic interface for different bot communication structs
type IBotAdapter interface {
	initialize()
}

// Admin is a structure for sirvice messages
type Admin struct {
}

// Recipient returns chatID for service messages
func (a *Admin) Recipient() string {
	return os.Getenv("ADMIN_CHAT_ID")
}

var db *sql.DB
var rootDir = "data"
var bot IBot

func main() {
	if os.Getenv("WORKING_DIR") != "" {
		rootDir = os.Getenv("WORKING_DIR")
	}

	logPath := path.Join(rootDir, "log.txt")
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("pullanusbot", true, false, lf).Close()
	// logger.Init("pullanusbot", true, true, lf)
	setupDB(rootDir)
	setupBot(os.Getenv("BOT_TOKEN"))

	adapters := []IBotAdapter{
		&Converter{},
		&Faggot{},
		&Info{},
		&SMS{},
		&TextHandler{handlers: []ITextHandler{
			&PlainLink{},
			&Twitter{},
		}},
		&Vpn{},
	}

	for _, adapter := range adapters {
		adapter.initialize()
	}

	bot.Start()
}

func setupDB(dir string) {
	logger.Info("Database initialization")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Warning("Directory not exist! Creating directory:")
		logger.Warning("\t" + dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			logger.Fatalf("Can't create directory: %s", dir)
		}
	}

	dbFile := path.Join(dir, "pullanusbot.db")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		logger.Warning("Database not exist! Creating database:")
		logger.Warning("\t" + dbFile)
		_, err = os.Create(dbFile)
		if err != nil {
			logger.Fatalf("Can't create database: %s", dbFile)
		}
	}

	db, _ = sql.Open("sqlite3", dbFile)

	logger.Info("Using database:")
	logger.Info("\t" + dbFile)
}

func setupBot(token string) {
	if token == "" {
		logger.Fatal("BOT_TOKEN not set")
	}

	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	var err error
	bot, err = tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		logger.Fatal(err)
	}
}
