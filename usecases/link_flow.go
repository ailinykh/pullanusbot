package usecases

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateLinkFlow is a basic LinkFlow factory
func CreateLinkFlow(l core.ILogger, hc core.IHttpClient, mf core.IMediaFactory, fd core.IFileDownloader, vff core.IVideoFactory, vfc core.IVideoConverter) *LinkFlow {
	return &LinkFlow{l, hc, mf, fd, vff, vfc}
}

// LinkFlow represents convert hotlink to video file logic
type LinkFlow struct {
	l   core.ILogger
	hc  core.IHttpClient
	mf  core.IMediaFactory
	fd  core.IFileDownloader
	vff core.IVideoFactory
	vfc core.IVideoConverter
}

// HandleText is a core.ITextHandler protocol implementation
func (lf *LinkFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`^http(\S+)$`)
	if r.MatchString(message.Text) {
		return lf.handleURL(message, bot)
	}
	return nil
}

func (lf *LinkFlow) handleURL(message *core.Message, bot core.IBot) error {
	contentType, err := lf.hc.GetContentType(message.Text)
	if err != nil {
		lf.l.Error(err)
		return err
	}

	if !strings.HasPrefix(contentType, "video") && !strings.HasPrefix(contentType, "image") {
		return nil
	}

	media, err := lf.mf.CreateMedia(message.Text, message.Sender)
	if err != nil {
		lf.l.Error(err)
		return err
	}

	for _, m := range media {
		switch m.Type {
		case core.TPhoto:
			err := lf.sendAsPhoto(m, message, bot)
			if err != nil {
				return err
			}
		case core.TVideo:
			err := lf.sendAsVideo(m, message, bot)
			if err != nil {
				return err
			}
		case core.TText:
			lf.l.Warningf("Unexpected %+v", m)
		}
	}

	return bot.Delete(message)
}

func (lf *LinkFlow) sendAsPhoto(media *core.Media, message *core.Message, bot core.IBot) error {
	lf.l.Infof("sending as photo: %s", media.URL)
	media.Caption = fmt.Sprintf(`<a href="%s">🖼</a> <b>%s</b> <i>(by %s)</i>`, media.URL, path.Base(media.URL), message.Sender.Username)
	_, err := bot.SendMedia(media)
	if err != nil {
		lf.l.Error(err)
		if strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") {
			file, err := lf.downloadMedia(media)
			if err != nil {
				return err
			}
			image := &core.Image{File: *file}
			_, err = bot.SendImage(image, media.Caption)
			return err
		}
	}
	return err
}

func (lf *LinkFlow) sendAsVideo(media *core.Media, message *core.Message, bot core.IBot) error {
	lf.l.Infof("sending as video: %s", media.URL)
	media.Caption = fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, message.Text, path.Base(media.URL), message.Sender.Username)

	if media.Codec != "h264" {
		lf.l.Warningf("expected h264 codec, but got %s", media.Codec)
		return lf.sendByConverting(media, bot)
	}

	_, err := bot.SendMedia(media)
	if err != nil {
		lf.l.Errorf("%s, fallback to uploading...", err)
		return lf.sendByUploading(media, bot)
	}

	return err
}

func (lf *LinkFlow) sendByConverting(media *core.Media, bot core.IBot) error {
	lf.l.Info("sending by converting")
	file, err := lf.downloadMedia(media)
	if err != nil {
		return err
	}

	vf, err := lf.vff.CreateVideo(file.Path)
	if err != nil {
		lf.l.Errorf("can't create video file for %s, %v", file.Path, err)
		return err
	}
	defer vf.Dispose()

	vfc, err := lf.vfc.Convert(vf, 0)
	if err != nil {
		lf.l.Errorf("cant convert video file: %v", err)
		return err
	}
	defer vfc.Dispose()

	_, err = bot.SendVideo(vfc, media.Caption)
	return err
}

func (lf *LinkFlow) sendByUploading(media *core.Media, bot core.IBot) error {
	// Try to upload file to telegram
	lf.l.Info("sending by uploading")

	file, err := lf.downloadMedia(media)
	if err != nil {
		return err
	}

	vf, err := lf.vff.CreateVideo(file.Path)
	if err != nil {
		lf.l.Errorf("can't create video file for %s, %v", file.Path, err)
		return err
	}

	defer vf.Dispose()
	_, err = bot.SendVideo(vf, media.Caption)
	return err
}

func (lf *LinkFlow) downloadMedia(media *core.Media) (*core.File, error) {
	mediaPath := path.Join(os.TempDir(), path.Base(media.URL))
	file, err := lf.fd.Download(media.URL, mediaPath)
	if err != nil {
		lf.l.Errorf("video download error: %v", err)
		return nil, err
	}

	lf.l.Infof("file downloaded: %s %0.2fMB", file.Name, float64(file.Size)/1024/1024)

	return file, nil
}
