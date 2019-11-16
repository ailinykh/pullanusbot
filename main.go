package main

import (
	"database/sql"
	"log"
	"os"
	"path"
	"time"

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

type Admin struct {
}

func (a *Admin) Recipient() string {
	return os.Getenv("ADMIN_CHAT_ID")
}

var db *sql.DB
var workingDir = "data"

var bot IBot

func main() {
	if os.Getenv("WORKING_DIR") != "" {
		workingDir = os.Getenv("WORKING_DIR")
	}

	if os.Getenv("DEV") == "" {
		logfile, err := os.OpenFile(path.Join(workingDir, "log.txt"), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("error opening file: %v", err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	setupDB(workingDir)
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
	log.Println("Database initialization")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Println("Directory not exist! Creating directory:")
		log.Println("\t" + dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("Can't create directory: %s", dir)
		}
	}

	dbFile := path.Join(dir, "pullanusbot.db")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Println("Database not exist! Creating database:")
		log.Println("\t" + dbFile)
		_, err = os.Create(dbFile)
		if err != nil {
			log.Fatalf("Can't create database: %s", dbFile)
		}
	}

	db, _ = sql.Open("sqlite3", dbFile)

	log.Println("Using database:")
	log.Println("\t" + dbFile)
}

func setupBot(token string) {
	if token == "" {
		log.Fatal("BOT_TOKEN not set")
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
		log.Fatal(err)
	}
}
