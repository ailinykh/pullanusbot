package use_cases

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateLinkFlow(l core.ILogger, fd core.IFileDownloader, vff core.IVideoFileFactory, vfc core.IVideoFileConverter) *LinkFlow {
	return &LinkFlow{l, fd, vff, vfc}
}

type LinkFlow struct {
	l   core.ILogger
	fd  core.IFileDownloader
	vff core.IVideoFileFactory
	vfc core.IVideoFileConverter
}

func (lf *LinkFlow) HandleText(text string, author *core.User, bot core.IBot) error {
	r := regexp.MustCompile(`^http(\S+)$`)
	if r.MatchString(text) {
		return lf.processLink(text, author, bot)
	}
	return nil
}

func (lf *LinkFlow) processLink(link string, author *core.User, bot core.IBot) error {
	resp, err := http.Get(link)

	if err != nil {
		lf.l.Error(err)
		return err
	}

	media := &core.Media{URL: resp.Request.URL.String()}
	media.Caption = fmt.Sprintf(`<a href="%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, link, path.Base(resp.Request.URL.Path), author.Username)

	switch resp.Header["Content-Type"][0] {
	case "video/mp4":
		lf.l.Infof("found mp4 file %s", link)
		err := bot.SendVideo(media)

		if err != nil {
			lf.l.Errorf("%s. Fallback to uploading", err)
			return lf.sendByUploading(media, bot)
		}
	case "video/webm":
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

		return bot.SendVideoFile(vfc, media.Caption)

	default:
		lf.l.Warningf("Unsupported content type: %s", resp.Header["Content-Type"])
	}
	return nil
}

func (lf *LinkFlow) sendByUploading(media *core.Media, bot core.IBot) error {
	// Try to upload file to telegram
	lf.l.Info("Sending by uploading")

	vf, err := lf.downloadMedia(media)
	if err != nil {
		return err
	}
	defer vf.Dispose()
	return bot.SendVideoFile(vf, media.Caption)
}

func (lf *LinkFlow) downloadMedia(media *core.Media) (*core.VideoFile, error) {
	filename := path.Base(media.URL)
	filepath := path.Join(os.TempDir(), filename)

	err := lf.fd.Download(media.URL, filepath)
	if err != nil {
		lf.l.Errorf("video download error: %v", err)
		return nil, err
	}

	stat, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	lf.l.Infof("File downloaded: %s %0.2fMB", filename, float64(stat.Size())/1024/1024)

	vf, err := lf.vff.CreateVideoFile(filepath)
	if err != nil {
		lf.l.Errorf("Can't create video file for %s, %v", filepath, err)
		return nil, err
	}
	return vf, nil
}
