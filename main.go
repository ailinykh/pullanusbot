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

var db *sql.DB

func main() {
	if os.Getenv("DEV") == "" {
		logfile, err := os.OpenFile("data/log.txt", os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("error opening file: %v", err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	setupdb("data")

	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Println("BOT_TOKEN not set")
		return
	}

	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		log.Println(err)
		return
	}

	game := NewFaggotGame(bot)
	game.Start()

	bot.Start()
}

func setupdb(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Directory not exist! Creating directory: %s", dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("Can't create directory: %s", dir)
		}
	}

	dbFile := path.Join(dir, "pullanusbot.db")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Printf("Database not exist! Creating database: %s", dbFile)
		_, err = os.Create(dbFile)
		if err != nil {
			log.Fatalf("Can't create database: %s", dbFile)
		}
	}

	db, _ = sql.Open("sqlite3", dbFile)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
