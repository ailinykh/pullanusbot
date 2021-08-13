package usecases

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateLinkFlow is a basic LinkFlow factory
func CreateLinkFlow(l core.ILogger, fd core.IFileDownloader, vff core.IVideoFactory, vfc core.IVideoConverter) *LinkFlow {
	return &LinkFlow{l, fd, vff, vfc}
}

// LinkFlow represents convert hotlink to video file logic
type LinkFlow struct {
	l   core.ILogger
	fd  core.IFileDownloader
	vff core.IVideoFactory
	vfc core.IVideoConverter
}

// HandleText is a core.ITextHandler protocol implementation
func (lf *LinkFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`^http(\S+)$`)
	if r.MatchString(message.Text) {
		return lf.processLink(message, bot)
	}
	return nil
}

func (lf *LinkFlow) processLink(message *core.Message, bot core.IBot) error {
	resp, err := http.Get(message.Text)

	if err != nil {
		lf.l.Error(err)
		return err
	}

	media := &core.Media{URL: resp.Request.URL.String()}
	media.Caption = fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, message.Text, path.Base(resp.Request.URL.Path), message.Sender.Username)

	switch resp.Header["Content-Type"][0] {
	case "video/mp4":
		lf.l.Infof("found mp4 file %s", message.Text)

		codec := lf.vfc.GetCodec(media.URL)
		if codec != "h264" {
			lf.l.Warningf("expected h264 codec, but got %s", codec)
			err := lf.sendByConverting(media, bot)
			if err != nil {
				return err
			}
		} else {
			_, err = bot.SendMedia(media)
			if err != nil {
				lf.l.Errorf("%s. Fallback to uploading", err)
				err := lf.sendByUploading(media, bot)
				if err != nil {
					return err
				}
			}
		}
	case "video/webm":
		err := lf.sendByConverting(media, bot)
		if err != nil {
			return err
		}
	case "text/html; charset=utf-8":
		return nil
	default:
		lf.l.Warningf("Unsupported content type: %s", resp.Header["Content-Type"])
		return nil
	}
	return bot.Delete(message)
}

func (lf *LinkFlow) sendByConverting(media *core.Media, bot core.IBot) error {
	vf, err := lf.downloadMedia(media)
	if err != nil {
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
	lf.l.Info("Sending by uploading")

	vf, err := lf.downloadMedia(media)
	if err != nil {
		return err
	}
	defer vf.Dispose()
	_, err = bot.SendVideo(vf, media.Caption)
	return err
}

func (lf *LinkFlow) downloadMedia(media *core.Media) (*core.Video, error) {
	mediaPath := path.Join(os.TempDir(), path.Base(media.URL))
	file, err := lf.fd.Download(media.URL, mediaPath)
	if err != nil {
		lf.l.Errorf("video download error: %v", err)
		return nil, err
	}

	lf.l.Infof("File downloaded: %s %0.2fMB", file.Name, file.Size/1024/1024)

	vf, err := lf.vff.CreateVideo(file.Path)
	if err != nil {
		lf.l.Errorf("Can't create video file for %s, %v", file.Path, err)
		return nil, err
	}
	return vf, nil
}
