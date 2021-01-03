package twitter

import (
	"fmt"
	"math"
	"os"
	"path"
	c "pullanusbot/converter"
	i "pullanusbot/interfaces"
	u "pullanusbot/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

var helper IHelper = Helper{}

// Twitter ...
type Twitter struct {
	bot     i.Bot
	matcher *regexp.Regexp
}

/*
Ну так-то ничего вопиющего не вижу. Но по-хорошему надо разобраться с обработкой ошибок и логированием.

Я бы не занимался обработкой ошибок в виде логирования в каждом методе, а оборачивал бы чем-то навроде
как в этом мануале https://medium.com/nuances-of-programming/b0db0c2131e8 - то есть кроме непосредственных
сообщений добавлял бы параметры и типизацию ошибки. И в итоге метод HandleTextMessage так же возвращал бы
ошибку вызывающему. И там уже можно было бы логировать или как-то иначе обрабатывать ошибки. Это может
пригодиться, например, если кроме бота этот функционал будет привязан например к REST где надо будет
возвращать ошибку юзеру, а не просто логировать.

По поводу логирования - по-хорошему в метод Setup нужно пробрасывать указатель на настроенный логгер.
И я бы склонился к https://godoc.org/github.com/sirupsen/logrus - он хорошо себя зарекомендовал.
Профит в том, что ты можешь пробросить какие-то кастомные парметры для логирования, чтобы потом
было проще найти кто высрал сообщение. Ну типа:

twitterLogger := logrus.WithField("serviceName", "twitter")

и потом например инициализируешь twitter.Setup(bot, twitterLogger)
И везде потом этот сервис будет срать указывая свой serviceName.

Если ошибки не высирать в лог прям на месте, а отправлять выше - то в коде остаются только логирования
уровня info или debug. И при случае можно будет включить логи только для конкретного сервиса, передав ему
инстанс с более низким LogLevel, оставив для всего приложения целиком какой-нибудь Warning level - будет
легче дебажить.

Вообще надо бы конечно мне это всё закнотрибутить, может осмелюсь на выходных как раз.
*/

// Setup all nesessary command handlers
func (t *Twitter) Setup(b i.Bot) {
	// Выпилим из глобальной видимости в инстанс, иначе каждая инициализация переписывает переменную bot
	t.bot = b
	// Компиляция регулярки это дорогая операция, сделаем ее при инициализации, а не на каждый запрос
	t.matcher, _ = regexp.Compile(`twitter\.com.+/(\d+)\S*$`)
	logger.Info("successfully initialized")
}

// HandleTextMessage is an i.TextMessageHandler interface implementation
func (t *Twitter) HandleTextMessage(m *tb.Message) {
	match := t.matcher.FindStringSubmatch(m.Text)
	if len(match) > 1 {
		t.processTweet(match[1], m)
	}
}

func (t *Twitter) processTweet(tweetID string, m *tb.Message) {
	logger.Infof("Processing tweet %s", tweetID)

	tweet, err := helper.getTweet(tweetID)
	if err != nil {
		return
	}

	if len(tweet.Errors) > 0 {
		tweet.ID = tweetID // In case of timeout
		t.processError(tweet, m)
		return
	}

	if len(tweet.ExtendedEntities.Media) == 0 && tweet.QuotedStatus != nil && len(tweet.QuotedStatus.ExtendedEntities.Media) > 0 {
		tweet = *tweet.QuotedStatus
		logger.Warningf("tweet media is empty, using QuotedStatus instead %s", tweet.ID)
	}

	media := tweet.ExtendedEntities.Media

	switch len(media) {
	case 0:
		t.sendText(tweet, m)
	case 1:
		switch media[0].Type {
		case "video", "animated_gif":
			t.sendVideo(media[0], tweet, m)
		case "photo":
			t.sendPhoto(media[0], tweet, m)
		default:
			// Вопиюще
			logger.Errorf("Unknown type: %s", media[0].Type)
			t.bot.Send(m.Chat, fmt.Sprintf("Unknown type: %s", media[0].Type), &tb.SendOptions{ReplyTo: m})
		}
	default:
		t.sendAlbum(media, tweet, m)
	}
}

func (t *Twitter) processError(tweet Tweet, m *tb.Message) {
	if tweet.Errors[0].Code == 88 { // "Rate limit exceeded 88"
		limit, err := strconv.ParseInt(tweet.Header["X-Rate-Limit-Reset"][0], 10, 64)
		if err != nil {
			logger.Error(err)
			return
		}
		timeout := limit - time.Now().Unix()
		logger.Infof("Twitter api timeout %d seconds", timeout)
		timeout = int64(math.Max(float64(timeout), 1)) // Twitter api timeout might be negative
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			t.processTweet(tweet.ID, m)
		}()
		return
	}

	logger.Error(tweet.Errors)
	logger.Info(tweet.Header)
	t.bot.Send(m.Chat, fmt.Sprint(tweet.Errors), &tb.SendOptions{ReplyTo: m})
}

func (t *Twitter) sendText(tweet Tweet, m *tb.Message) {
	logger.Info("Sending as text")
	caption := helper.makeCaption(m, tweet)
	for _, url := range tweet.Entities.Urls {
		caption += "\n" + url.ExpandedURL
	}
	_, err := t.bot.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true})
	if err != nil {
		logger.Error(err)
		return
	}
	t.deleteMessage(m)
}

func (t *Twitter) sendPhoto(media Media, tweet Tweet, m *tb.Message) {
	file := &tb.Photo{File: tb.FromURL(media.MediaURL)}
	file.Caption = helper.makeCaption(m, tweet)
	logger.Infof("Sending as Photo %s", file.FileURL)
	t.bot.Notify(m.Chat, tb.UploadingPhoto)
	_, err := t.bot.Send(m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		logger.Error(err)
		return
	}
	t.deleteMessage(m)
}

func (t *Twitter) sendAlbum(media []Media, tweet Tweet, m *tb.Message) {
	logger.Infof("Sending as Album")
	caption := helper.makeCaption(m, tweet)
	t.bot.Notify(m.Chat, tb.UploadingPhoto)
	_, err := t.bot.SendAlbum(m.Chat, helper.makeAlbum(media, caption))
	if err != nil {
		logger.Error(err)
		return
	}
	t.deleteMessage(m)
}

func (t *Twitter) sendVideo(media Media, tweet Tweet, m *tb.Message) {
	file := &tb.Video{File: tb.FromURL(media.VideoInfo.best().URL)}
	caption := helper.makeCaption(m, tweet)
	file.Caption = caption
	logger.Infof("Sending as Video %s", file.FileURL)
	t.bot.Notify(m.Chat, tb.UploadingVideo)
	_, err := t.bot.Send(m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err == nil {
		t.deleteMessage(m)
		return
	}
	if !(strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified")) {
		logger.Error(err)
		return
	}
	// Try to upload file to telegram
	logger.Info("Sending by uploading")

	filename := path.Base(media.VideoInfo.best().URL)
	filepath := path.Join(os.TempDir(), filename)
	defer os.Remove(filepath)

	err = u.DownloadFile(filepath, media.VideoInfo.best().URL)
	if err != nil {
		logger.Errorf("video download error: %v", err)
		return
	}

	videofile, err := c.NewVideoFile(filepath)
	if err != nil {
		logger.Errorf("Can't create video file for %s, %v", filepath, err)
		return
	}
	defer videofile.Dispose()
	videofile.Upload(t.bot, m, caption, c.UploadFinishedCallback)
}

func (t *Twitter) deleteMessage(m *tb.Message) {
	err := t.bot.Delete(m)
	if err != nil {
		logger.Warningf("Can't delete original message: %v", err)
	}
}
