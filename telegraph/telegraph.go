package telegraph

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	i "pullanusbot/interfaces"
	"pullanusbot/utils"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

type telegraphImage struct {
	Src string `json:"src"`
}

// Telegraph upload images to telegra.ph
type Telegraph struct {
}

// Setup all nesessary command handlers
func (t *Telegraph) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	bot.Handle(tb.OnPhoto, t.upload)
	logger.Info("Successfully initialized")
}

func (t *Telegraph) upload(m *tb.Message) {
	if !m.Private() {
		return
	}

	tmpFile := path.Join(os.TempDir(), "tmp-image-"+utils.RandStringRunes(4)+".jpg")
	defer os.Remove(tmpFile)

	err := bot.Download(&m.Photo.File, tmpFile)

	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(tmpFile))
	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}

	req, err := http.NewRequest("POST", "https://telegra.ph/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}

	body2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}

	var images []telegraphImage
	err = json.Unmarshal(body2, &images)
	if err != nil {
		logger.Error(err)
		bot.Send(m.Chat, err, &tb.SendOptions{ReplyTo: m})
		return
	}

	url := "https://telegra.ph" + images[0].Src
	logger.Infof("%d %s", m.Chat.ID, url)
	bot.Send(m.Chat, url, &tb.SendOptions{DisableWebPagePreview: true})
}
