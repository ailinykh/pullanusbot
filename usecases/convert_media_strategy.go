package usecases

import (
	"os"
	"path"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateConvertMediaStrategy(l core.ILogger, sms core.ISendMediaStrategy, fd core.IFileDownloader, vf core.IVideoFactory, vc core.IVideoConverter) *ConvertMediaStrategy {
	return &ConvertMediaStrategy{l, sms, fd, vf, vc}
}

type ConvertMediaStrategy struct {
	l   core.ILogger
	sms core.ISendMediaStrategy
	fd  core.IFileDownloader
	vf  core.IVideoFactory
	vc  core.IVideoConverter
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (cms *ConvertMediaStrategy) SendMedia(media []*core.Media, bot core.IBot) error {
	for _, m := range media {
		if m.Type == core.TVideo && media[0].Codec != "mp4" {
			return cms.fallbackToConverting(m, bot)
		}
	}
	return cms.sms.SendMedia(media, bot)
}

func (cms *ConvertMediaStrategy) fallbackToConverting(media *core.Media, bot core.IBot) error {
	cms.l.Info("send by converting")
	file, err := cms.downloadMedia(media)
	if err != nil {
		return err
	}
	defer file.Dispose()

	vf, err := cms.vf.CreateVideo(file.Path)
	if err != nil {
		cms.l.Errorf("can't create video file for %s, %v", file.Path, err)
		return err
	}
	defer vf.Dispose()

	vfc, err := cms.vc.Convert(vf, 0)
	if err != nil {
		cms.l.Errorf("cant convert video file: %v", err)
		return err
	}
	defer vfc.Dispose()

	_, err = bot.SendVideo(vfc, media.Caption)
	return err
}

func (cms *ConvertMediaStrategy) downloadMedia(media *core.Media) (*core.File, error) {
	mediaPath := path.Join(os.TempDir(), path.Base(media.URL))
	file, err := cms.fd.Download(media.URL, mediaPath)
	if err != nil {
		cms.l.Errorf("video download error: %v", err)
		return nil, err
	}

	cms.l.Infof("file downloaded: %s %0.2fMB", file.Name, float64(file.Size)/1024/1024)

	return file, nil
}
