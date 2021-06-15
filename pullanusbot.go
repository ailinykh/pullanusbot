package main

import (
	"math/rand"
	"os"
	"path"
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

	telebot := api.CreateTelebot(os.Getenv("BOT_TOKEN"), logger)

	localizer := infrastructure.GameLocalizer{}
	game := use_cases.CreateGameFlow(localizer)
	telebot.SetupGame(game)

	telebot.SetupInfo()

	converter := infrastructure.CreateFfmpegConverter()
	video_flow := use_cases.CreateVideoFlow(logger, converter, converter)
	telebot.AddHandler(video_flow)

	file_downloader := infrastructure.CreateFileDownloader()
	twitter_api := api.CreateTwitterAPI()
	twitter_flow := use_cases.CreateTwitterFlow(logger, twitter_api, file_downloader, converter)
	telebot.AddHandler(twitter_flow)

	link_flow := use_cases.CreateLinkFlow(logger, file_downloader, converter, converter)
	telebot.AddHandler(link_flow)

	file_uploader := api.CreateTelegraphAPI()
	//TODO: image_downloader := api.CreateTelebotImageDownloader()
	image_flow := use_cases.CreateImageFlow(logger, file_uploader, telebot)
	telebot.AddHandler(image_flow)

	publisher_flow := use_cases.CreatePublisherFlow(logger)
	telebot.AddHandler(publisher_flow)
	telebot.AddHandler("/loh666", publisher_flow)
	// Start endless loop
	telebot.Run()
}

func createLogger() (core.ILogger, func()) {
	logFilePath := path.Join(getWorkingDir(), "pullanusbot.log")
	lf, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
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

func getWorkingDir() string {
	workingDir := os.Getenv("WORKING_DIR")
	if len(workingDir) == 0 {
		return "pullanusbot-data"
	}
	return workingDir
}
