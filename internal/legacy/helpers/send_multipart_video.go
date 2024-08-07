package helpers

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

// FIXME: SendMultipartVideo should conform to core.ISendVideoStrategy
func CreateSendMultipartVideo(l core.ILogger, url core.URL) *SendMultipartVideo {
	return &SendMultipartVideo{l, http.DefaultClient, url}
}

type SendMultipartVideo struct {
	l      core.ILogger
	client *http.Client
	url    core.URL
}

func (strategy *SendMultipartVideo) SendVideo(video *core.Video, caption string, chatId int64) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	strategy.addParams(writer, map[string]interface{}{
		"caption":            caption,
		"duration":           video.Duration,
		"width":              video.Width,
		"height":             video.Height,
		"supports_streaming": "true",
		"parse_mode":         "HTML",
		"chat_id":            chatId,
		"video":              video.File,
		"thumb":              video.Thumb.File,
	})

	writer.Close()

	start := time.Now()
	strategy.l.Infof("uploading %s (%.2f MB)", video.Name, float64(video.Size)/1024/1024)
	r, _ := http.NewRequest("POST", strategy.url, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := strategy.client.Do(r)
	if err != nil {
		strategy.l.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	strategy.l.Infof("successfully sent %s (%.2f MB) %s", video.Name, float64(video.Size)/1024/1024, time.Now().Sub(start))
	return ioutil.ReadAll(res.Body)
}

func (strategy *SendMultipartVideo) addParams(writer *multipart.Writer, params map[string]interface{}) {
	for key, param := range params {
		var reader io.Reader
		var part io.Writer
		var err error
		switch p := param.(type) {
		case string:
			part, err = writer.CreateFormField(key)
			reader = strings.NewReader(p)
		case int:
			part, err = writer.CreateFormField(key)
			reader = strings.NewReader(strconv.Itoa(p))
		case int64:
			part, err = writer.CreateFormField(key)
			reader = strings.NewReader(strconv.FormatInt(p, 10))
		case core.File:
			file, err := os.Open(p.Path)
			if err != nil {
				strategy.l.Error(err)
				continue
			}
			defer file.Close()
			part, err = writer.CreateFormFile(key, file.Name())
			reader = file
		default:
			strategy.l.Errorf("unexpected param type %+v", p)
			continue
		}

		if err != nil {
			strategy.l.Error(err)
			continue
		}
		_, err = io.Copy(part, reader)
		if err != nil {
			strategy.l.Error(err)
		}
	}
}
