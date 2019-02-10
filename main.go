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
var workingDir = "data"

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

	setupdb(workingDir)
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

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
