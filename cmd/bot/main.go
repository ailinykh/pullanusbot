package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/ailinykh/pullanusbot/v2/internal/api/logger"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/api"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/helpers"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/usecases"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	config := NewDefaultConfig()
	logger := logger.NewGoogleLogger(ctx, config.WorkingDir())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Error("signal received", "signal", sig)
		cancel()
	}()

	dbFile := path.Join(config.WorkingDir(), "pullanusbot.db")

	settingsProvider := infrastructure.CreateSettingsStorage(dbFile)
	boolSettingProvider := helpers.CreateBoolSettingProvider(settingsProvider)
	databaseChatStorage := infrastructure.CreateChatStorage(dbFile, logger)
	inMemoryChatStorage := infrastructure.CreateInMemoryChatStorage()
	chatStorageDecorator := usecases.CreateChatStorageDecorator(inMemoryChatStorage, databaseChatStorage)

	chatID, err := strconv.ParseInt(os.Getenv("ADMIN_CHAT_ID"), 10, 64)
	if err != nil {
		chatID = 0
	}
	telebot := api.CreateTelebot(
		logger,
		api.WithBotToken(config.BotToken()),
		api.WithReportChatId(chatID),
	)
	telebot.SetupInfo()

	databaseUserStorage := infrastructure.CreateUserStorage(dbFile, logger)
	inMemoryUserStorage := infrastructure.CreateInMemoryUserStorage()
	userStorageDecorator := usecases.CreateUserStorageDecorator(inMemoryUserStorage, databaseUserStorage)
	bootstrapFlow := usecases.CreateBootstrapFlow(logger, chatStorageDecorator, userStorageDecorator)
	telebot.AddHandler(bootstrapFlow)

	localizer := infrastructure.CreateGameLocalizer()
	gameStorage := infrastructure.CreateGameStorage(dbFile)
	rand := infrastructure.CreateMathRand()
	commandService := usecases.CreateCommandService()
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
	remoteMediaSender := helpers.CreateSendMediaStrategy()
	sendVideoStrategy := helpers.CreateSendVideoStrategy()
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
	ytdlpApi := api.CreateYtDlpApi([]string{}, logger)
	tiktokMediaFactory := api.CreateTikTokMediaFactory(ytdlpApi)
	tiktokFlow := usecases.CreateTikTokFlow(tiktokHttpClient, tiktokMediaFactory, localMediaSender)
	telebot.AddHandler(tiktokFlow)

	fileUploader := api.CreateTelegraphAPI()
	//TODO: image_downloader := api.CreateTelebotImageDownloader()
	imageFlow := usecases.CreateImageFlow(logger, fileUploader, telebot)
	telebot.AddHandler(imageFlow)

	{
		chatId := os.Getenv("PUBLISHER_CHAT_ID")
		username := os.Getenv("PUBLISHER_USERNAME")
		if len(chatId) > 0 && len(username) > 0 {
			logger.Info("publisher logic enabled", "chat_id", chatId, "username", username)
			chatID, err := strconv.ParseInt(chatId, 10, 64)
			if err != nil {
				logger.Error("failed to parse publisher chat id", "error", err)
			} else {
				publisherFlow := usecases.CreatePublisherFlow(chatID, username, logger)
				telebot.AddHandler(publisherFlow)
				telebot.AddHandler("/loh666", publisherFlow.HandleRequest)
			}
		}
	}

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
		keyId := os.Getenv("LIGHTSAIL_ACCESS_KEY_ID")
		secret := os.Getenv("LIGHTSAIL_SECRET_ACCESS_KEY")
		chatId := os.Getenv("REBOOT_SERVER_CHAT_ID")
		command := os.Getenv("REBOOT_SERVER_COMMAND")
		if len(keyId) > 0 && len(secret) > 0 && len(chatId) > 0 && len(command) > 0 {
			chatIds := []int64{}
			for _, chatId := range strings.Split(chatId, ",") {
				chatID, err := strconv.ParseInt(chatId, 10, 64)
				if err != nil {
					logger.Error("failed to parse reboot server chat id", "error", err)
				} else {
					chatIds = append(chatIds, chatID)
				}
			}
			if len(chatIds) > 0 {
				logger.Info("server reboot logic enabled", "chats", chatIds, "command", command)
				lightsailApi := api.NewLightsailAPI(logger, keyId, secret)
				opts := &usecases.RebootServerOptions{
					ChatIds: chatIds,
					Command: command,
				}
				rebootFlow := usecases.NewRebootServerFlow(lightsailApi, commandService, logger, opts)
				telebot.AddHandler(rebootFlow)
			} else {
				logger.Warn("server reboot logic disabled due to no chat id's specified")
			}
		} else {
			logger.Info("server reboot logic disabled")
		}
	}

	iDoNotCare := usecases.CreateIDoNotCare()
	telebot.AddHandler(iDoNotCare)

	if cookiesFilePath := os.Getenv("INSTAGRAM_COOKIES_FILE_PATH"); len(cookiesFilePath) > 0 {
		logger.Info("instagram logic enabled. Cookies file: %s", cookiesFilePath)
		cookies := path.Join(config.WorkingDir(), cookiesFilePath)
		instaAPI := api.CreateYtDlpApi([]string{"--cookies", cookies}, logger)
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
	go telebot.Run()

	logger.Info("waiting for context...")
	<-ctx.Done()
	logger.Info("attempt to shutdown gracefully...")
}
