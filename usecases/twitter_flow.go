package usecases

import (
	"errors"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterFlow(l core.ILogger, mf core.IMediaFactory, fd core.IFileDownloader, vff core.IVideoFactory) *TwitterFlow {
	return &TwitterFlow{l, mf, fd, vff}
}

// TwitterFlow represents tweet processing logic
type TwitterFlow struct {
	l   core.ILogger
	mf  core.IMediaFactory
	fd  core.IFileDownloader
	vff core.IVideoFactory
}

// HandleText is a core.ITextHandler protocol implementation
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
	media, err := tf.mf.CreateMedia(tweetID, message.Sender)
	if err != nil {
		tf.l.Error(err)
		return err
	}

	err = tf.handleMedia(media, message, bot)
	if err != nil {
		tf.l.Error(err)
		return err
	}

	return bot.Delete(message)
}

func (tf *TwitterFlow) handleMedia(media []*core.Media, message *core.Message, bot core.IBot) error {
	switch len(media) {
	case 0:
		return errors.New("unexpected 0 media count")
	case 1:
		_, err := bot.SendMedia(media[0])
		if err != nil {
			if strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") {
				return tf.fallbackToUploading(media[0], bot)
			}
		}
		return err
	default:
		_, err := bot.SendPhotoAlbum(media)
		return err
	}
}

func (tf *TwitterFlow) fallbackToUploading(media *core.Media, bot core.IBot) error {
	// Try to upload file to telegram
	tf.l.Info("Sending by uploading")
	mediaPath := path.Join(os.TempDir(), path.Base(media.URL))
	file, err := tf.fd.Download(media.URL, mediaPath)
	if err != nil {
		tf.l.Errorf("file download error: %v", err)
		return err
	}

	defer file.Dispose()

	stat, err := os.Stat(file.Path)
	if err != nil {
		return err
	}

	tf.l.Infof("File downloaded: %s %0.2fMB", file.Name, float64(stat.Size())/1024/1024)

	if media.Type == core.TPhoto {
		image := &core.Image{File: *file}
		_, err := bot.SendImage(image, media.Caption)
		return err
	}
	// else
	vf, err := tf.vff.CreateVideo(file.Path)
	if err != nil {
		tf.l.Errorf("Can't create video file for %s, %v", file.Path, err)
		return err
	}
	defer vf.Dispose()
	_, err = bot.SendVideo(vf, media.Caption)
	return err
}
