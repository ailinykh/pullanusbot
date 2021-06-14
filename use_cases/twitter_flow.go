package use_cases

import (
	"errors"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterFlow(l core.ILogger, mf core.IMediaFactory, fd core.IFileDownloader, vff core.IVideoFileFactory) *TwitterFlow {
	return &TwitterFlow{l, mf, fd, vff}
}

type TwitterFlow struct {
	l   core.ILogger
	mf  core.IMediaFactory
	fd  core.IFileDownloader
	vff core.IVideoFileFactory
}

// core.ITextHandler
func (tf *TwitterFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`twitter\.com.+/(\d+)\S*$`)
	match := r.FindStringSubmatch(message.Text)
	if len(match) < 2 {
		return nil // no tweet id found
	}
	return tf.process(match[1], message, bot)
}

func (tf *TwitterFlow) process(tweetID string, message *core.Message, bot core.IBot) error {
	tf.l.Infof("processing tweet %s", tweetID)
	medias, err := tf.mf.CreateMedia(tweetID, message.Sender)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Rate limit exceeded") {
			return tf.handleTimeout(err, tweetID, message, bot)
		}
		return err
	}

	switch len(medias) {
	case 0:
		return errors.New("unexpected 0 media count")
	case 1:
		_, err := bot.SendMedia(medias[0])
		if err != nil && medias[0].Type == core.Video {
			if strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") {
				return tf.sendByUploading(medias[0], bot)
			}
		}
		return err
	default:
		_, err := bot.SendPhotoAlbum(medias)
		return err
	}
}

func (tf *TwitterFlow) handleTimeout(err error, tweetID string, message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`(\-?\d+)$`)
	match := r.FindStringSubmatch(err.Error())
	if len(match) < 2 {
		return errors.New("rate limit not found")
	}

	limit, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return err
	}

	timeout := limit - time.Now().Unix()
	tf.l.Infof("Twitter api timeout %d seconds", timeout)
	timeout = int64(math.Max(float64(timeout), 1)) // Twitter api timeout might be negative
	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		tf.process(tweetID, message, bot)
	}()
	return nil // TODO: is it ok?
}

func (tf *TwitterFlow) sendByUploading(media *core.Media, bot core.IBot) error {
	// Try to upload file to telegram
	tf.l.Info("Sending by uploading")
	file, err := tf.fd.Download(media.URL)
	if err != nil {
		tf.l.Errorf("video download error: %v", err)
		return err
	}

	defer os.Remove(file.Path)

	stat, err := os.Stat(file.Path)
	if err != nil {
		return err
	}

	tf.l.Infof("File downloaded: %s %0.2fMB", file.Name, float64(stat.Size())/1024/1024)

	vf, err := tf.vff.CreateVideoFile(file.Path)
	if err != nil {
		tf.l.Errorf("Can't create video file for %s, %v", file.Path, err)
		return err
	}
	defer vf.Dispose()
	_, err = bot.SendVideoFile(vf, media.Caption)
	return err
}
