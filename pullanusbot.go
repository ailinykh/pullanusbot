package main

import (
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/ailinykh/pullanusbot/v2/api"
	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"github.com/google/logger"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logger, close := createLogger()
	defer close()

	telebot := api.CreateTelebot(os.Getenv("BOT_TOKEN"), logger)
	telebot.SetupInfo()

	localizer := infrastructure.GameLocalizer{}
	dbFile := path.Join(getWorkingDir(), "pullanusbot.db")
	gameStorage := infrastructure.CreateGameStorage(dbFile)
	gameFlow := usecases.CreateGameFlow(localizer, gameStorage)
	telebot.AddHandler("/pidorules", gameFlow.Rules)
	telebot.AddHandler("/pidoreg", gameFlow.Add)
	telebot.AddHandler("/pidor", gameFlow.Play)
	telebot.AddHandler("/pidorstats", gameFlow.Stats)
	telebot.AddHandler("/pidorall", gameFlow.All)
	telebot.AddHandler("/pidorme", gameFlow.Me)

	converter := infrastructure.CreateFfmpegConverter(logger)
	videoFlow := usecases.CreateVideoFlow(logger, converter, converter)
	telebot.AddHandler(videoFlow)

	fileDownloader := infrastructure.CreateFileDownloader()
	remoteMediaSender := usecases.CreateSendMediaStrategy(logger)
	localMediaSender := usecases.CreateUploadMediaStrategy(logger, remoteMediaSender, fileDownloader, converter, converter)
	twitterAPI := api.CreateTwitterAPI()
	twitterFlow := usecases.CreateTwitterFlow(logger, twitterAPI, localMediaSender)
	twitterTimeout := usecases.CreateTwitterTimeout(logger, twitterFlow)
	twitterParser := usecases.CreateTwitterParser(twitterTimeout)
	telebot.AddHandler(twitterParser)

	httpClient := api.CreateHttpClient()
	mp4MediaSender := usecases.CreateConvertMediaStrategy(logger, localMediaSender, fileDownloader, converter, converter)
	linkFlow := usecases.CreateLinkFlow(logger, httpClient, converter, mp4MediaSender)
	telebot.AddHandler(linkFlow)

	fileUploader := api.CreateTelegraphAPI()
	//TODO: image_downloader := api.CreateTelebotImageDownloader()
	imageFlow := usecases.CreateImageFlow(logger, fileUploader, telebot)
	telebot.AddHandler(imageFlow)

	publisherFlow := usecases.CreatePublisherFlow(logger)
	telebot.AddHandler(publisherFlow)
	telebot.AddHandler("/loh666", publisherFlow.HandleRequest)

	youtubeAPI := api.CreateYoutubeAPI(logger, fileDownloader)
	youtubeFlow := usecases.CreateYoutubeFlow(logger, youtubeAPI, youtubeAPI, converter)
	telebot.AddHandler(youtubeFlow)

	telebot.AddHandler("/proxy", func(m *core.Message, bot core.IBot) error {
		_, err := bot.SendText("tg://proxy?server=proxy.ailinykh.com&port=443&secret=dd71ce3b5bf1b7015dc62a76dc244c5aec")
		return err
	})

	iDoNotCare := usecases.CreateIDoNotCare()
	telebot.AddHandler(iDoNotCare)
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
