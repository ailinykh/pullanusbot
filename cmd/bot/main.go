package main

import (
	"os"
	"path"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/api"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/helpers"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/usecases"
	"github.com/google/logger"
)

func main() {
	config := NewDefaultConfig()
	logger := createLogger(config.WorkingDir())
	defer logger.Close()

	dbFile := path.Join(config.WorkingDir(), "pullanusbot.db")

	settingsProvider := infrastructure.CreateSettingsStorage(dbFile)
	boolSettingProvider := helpers.CreateBoolSettingProvider(settingsProvider)
	databaseChatStorage := infrastructure.CreateChatStorage(dbFile, logger)
	inMemoryChatStorage := infrastructure.CreateInMemoryChatStorage()
	chatStorageDecorator := usecases.CreateChatStorageDecorator(logger, inMemoryChatStorage, databaseChatStorage)
	telebot := api.CreateTelebot(config.BotToken(), logger)
	telebot.SetupInfo()

	databaseUserStorage := infrastructure.CreateUserStorage(dbFile, logger)
	inMemoryUserStorage := infrastructure.CreateInMemoryUserStorage()
	userStorageDecorator := usecases.CreateUserStorageDecorator(inMemoryUserStorage, databaseUserStorage)
	bootstrapFlow := usecases.CreateBootstrapFlow(logger, chatStorageDecorator, userStorageDecorator)
	telebot.AddHandler(bootstrapFlow)

	localizer := infrastructure.CreateGameLocalizer()
	gameStorage := infrastructure.CreateGameStorage(dbFile)
	rand := infrastructure.CreateMathRand()
	commandService := usecases.CreateCommandService(logger)
	gameFlow := usecases.CreateGameFlow(logger, localizer, gameStorage, rand, settingsProvider, commandService)
	telebot.AddHandler("/pidorules", gameFlow.Rules)
	telebot.AddHandler("/pidoreg", gameFlow.Add)
	telebot.AddHandler("/pidor", gameFlow.Play)
	telebot.AddHandler("/pidorstats", gameFlow.Stats)
	telebot.AddHandler("/pidorall", gameFlow.All)
	telebot.AddHandler("/pidorme", gameFlow.Me)

	converter := infrastructure.CreateFfmpegConverter(logger)
	videoFlow := usecases.CreateVideoFlow(logger, converter, converter)
	telebot.AddHandler(videoFlow)

	fileDownloader := infrastructure.CreateFileDownloader(logger)
	remoteMediaSender := helpers.CreateSendMediaStrategy(logger)
	sendVideoStrategy := helpers.CreateSendVideoStrategy(logger)
	sendVideoStrategySplitDecorator := helpers.CreateSendVideoStrategySplitDecorator(logger, sendVideoStrategy, converter)
	localMediaSender := helpers.CreateUploadMediaDecorator(logger, remoteMediaSender, fileDownloader, converter, sendVideoStrategySplitDecorator)

	rabbit, close := infrastructure.CreateRabbitFactory(logger, config.AmqpUrl())
	defer close()
	task := rabbit.NewTask("twitter_queue")

	twitterMediaFactory := api.CreateTwitterMediaFactory(logger, task)
	twitterFlow := usecases.CreateTwitterFlow(logger, twitterMediaFactory, localMediaSender)
	twitterTimeout := usecases.CreateTwitterTimeout(logger, twitterFlow)
	twitterParser := usecases.CreateTwitterParser(logger, twitterTimeout)
	twitterRemoveSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, twitterParser, core.STwitterFlowRemoveSource, boolSettingProvider)
	telebot.AddHandler(twitterRemoveSourceDecorator)

	httpClient := api.CreateHttpClient()
	convertMediaSender := helpers.CreateConvertMediaStrategy(logger, localMediaSender, fileDownloader, converter, converter)
	linkFlow := usecases.CreateLinkFlow(logger, httpClient, converter, convertMediaSender)
	removeLinkSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, linkFlow, core.SLinkFlowRemoveSource, boolSettingProvider)
	telebot.AddHandler(removeLinkSourceDecorator)

	tiktokHttpClient := api.CreateHttpClient() // domain specific headers and cookies
	ytdlpApi := api.CreateYtDlpApi("", logger)
	tiktokMediaFactory := api.CreateTikTokMediaFactory(logger, ytdlpApi)
	tiktokFlow := usecases.CreateTikTokFlow(logger, tiktokHttpClient, tiktokMediaFactory, localMediaSender)
	telebot.AddHandler(tiktokFlow)

	fileUploader := api.CreateTelegraphAPI()
	//TODO: image_downloader := api.CreateTelebotImageDownloader()
	imageFlow := usecases.CreateImageFlow(logger, fileUploader, telebot)
	telebot.AddHandler(imageFlow)

	publisherFlow := usecases.CreatePublisherFlow(logger)
	telebot.AddHandler(publisherFlow)
	telebot.AddHandler("/loh666", publisherFlow.HandleRequest)

	youtubeMediaFactory := api.CreateYoutubeMediaFactory(logger, ytdlpApi, fileDownloader)
	youtubeFlow := usecases.CreateYoutubeFlow(logger, youtubeMediaFactory, youtubeMediaFactory, sendVideoStrategySplitDecorator)
	removeYoutubeSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, youtubeFlow, core.SYoutubeFlowRemoveSource, boolSettingProvider)
	telebot.AddHandler(removeYoutubeSourceDecorator)

	telebot.AddHandler("/proxy", func(m *core.Message, bot core.IBot) error {
		_ = commandService.EnableCommands(m.Chat.ID, []core.Command{{Text: "proxy", Description: "proxy server for telegram"}}, bot)
		_, err := bot.SendText("tg://proxy?server=proxy.ailinykh.com&port=443&secret=dd71ce3b5bf1b7015dc62a76dc244c5aec")
		return err
	})

	{
		keyId := config.StringForKey("LIGHTSAIL_ACCESS_KEY_ID")
		secret := config.StringForKey("LIGHTSAIL_SECRET_ACCESS_KEY")
		chatId := config.StringForKey("REBOOT_SERVER_CHAT_ID")
		command := config.StringForKey("REBOOT_SERVER_COMMAND")
		if len(keyId) > 0 && len(secret) > 0 && len(chatId) > 0 && len(command) > 0 {
			logger.Infof("server reboot logic enabled for %s by %s", chatId, command)
			chatID, err := strconv.ParseInt(chatId, 10, 64)
			if err != nil {
				logger.Errorf("failed to parse %s: %v", chatID, err)
			} else {
				lightsailApi := api.NewLightsailAPI(logger, keyId, secret)
				opts := &usecases.RebootServerOptions{
					ChatId:  chatID,
					Command: command,
				}
				rebootFlow := usecases.NewRebootServerFlow(lightsailApi, commandService, logger, opts)
				telebot.AddHandler(rebootFlow)
			}
		} else {
			logger.Info("server reboot logic disabled")
		}
	}

	iDoNotCare := usecases.CreateIDoNotCare()
	telebot.AddHandler(iDoNotCare)

	if cookiesFilePath := config.StringForKey("INSTAGRAM_COOKIES_FILE_PATH"); len(cookiesFilePath) > 0 {
		logger.Infof("instagram logic enabled. Cookies file: %s", cookiesFilePath)
		instaAPI := api.CreateYtDlpApi(path.Join(config.WorkingDir(), cookiesFilePath), logger)
		instaFlow := usecases.CreateInstagramFlow(logger, instaAPI, localMediaSender)
		removeInstaSourceDecorator := usecases.CreateRemoveSourceDecorator(logger, instaFlow, core.SInstagramFlowRemoveSource, boolSettingProvider)
		telebot.AddHandler(removeInstaSourceDecorator)
	} else {
		logger.Info("instagram logic disabled")
	}

	commonLocalizer := infrastructure.CreateCommonLocalizer()
	startFlow := usecases.CreateStartFlow(logger, commonLocalizer, settingsProvider, commandService)
	telebot.AddHandler("/start", startFlow.Start)
	telebot.AddHandler("/help", startFlow.Help)
	// Start endless loop
	telebot.Run()
}

func createLogger(workingDir string) core.ILogger {
	logFilePath := path.Join(workingDir, "pullanusbot.log")
	lf, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	return logger.Init("pullanusbot", true, false, lf)
}
