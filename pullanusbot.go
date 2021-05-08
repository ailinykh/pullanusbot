package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/ailinykh/pullanusbot/v2/api"
	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/use_cases"
	"github.com/google/logger"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logger, close := createLogger()
	defer close()

	converter := infrastructure.CreateFfmpegConverter()
	telebot := api.CreateTelebot(os.Getenv("BOT_TOKEN"), logger, converter)

	localizer := infrastructure.GameLocalizer{}
	game := use_cases.CreateGameFlow(localizer)
	telebot.SetupGame(game)

	telebot.SetupInfo()

	video_flow := use_cases.CreateVideoFlow(logger, converter, converter)
	telebot.SetupVideo(video_flow)
	// Start endless loop
	telebot.Run()
}

func createLogger() (core.ILogger, func()) {
	lf, err := os.OpenFile("pullanusbot.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	l := logger.Init("pullanusbot", true, false, lf)
	close := func() {
		lf.Close()
		l.Close()
	}
	return l, close
}
