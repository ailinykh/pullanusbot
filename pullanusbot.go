package main

import (
	"math/rand"
	"time"

	"github.com/ailinykh/pullanusbot/v2/api"
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/use_cases"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	localizer := infrastructure.GameLocalizer{}
	game := use_cases.CreateGameFlow(localizer)
	telebot := api.CreateTelebot("TOKEN")

	telebot.SetupGame(game)
	telebot.SetupInfo()
	// Start endless loop
	telebot.Run()
}
