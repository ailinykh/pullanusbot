package usecases

import (
	"os"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateUploadMediaStrategy(l core.ILogger, sms core.ISendMediaStrategy, fd core.IFileDownloader, vf core.IVideoFactory, vc core.IVideoConverter) *UploadMediaStrategy {
	return &UploadMediaStrategy{l, sms, fd, vf, vc}
}

type UploadMediaStrategy struct {
	l   core.ILogger
	sms core.ISendMediaStrategy
	fd  core.IFileDownloader
	vf  core.IVideoFactory
	vc  core.IVideoConverter
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (ums *UploadMediaStrategy) SendMedia(media []*core.Media, bot core.IBot) error {
	err := ums.sms.SendMedia(media, bot)
	if err != nil {
		ums.l.Error(err)
		if strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") {
			return ums.fallbackToUploading(media[0], bot)
		}
	}

	return err
}

func (ums *UploadMediaStrategy) fallbackToUploading(media *core.Media, bot core.IBot) error {
	ums.l.Info("send by uploading")
	file, err := ums.downloadMedia(media)
	if err != nil {
		return err
	}
	defer file.Dispose()

	switch media.Type {
	case core.TText:
		ums.l.Warning("unexpected media type")
	case core.TPhoto:
		image := &core.Image{File: *file}
		_, err = bot.SendImage(image, media.Caption)
		return err
	case core.TVideo:
		vf, err := ums.vf.CreateVideo(file.Path)
		if err != nil {
			ums.l.Errorf("can't create video file for %s, %v", file.Path, err)
			return err
		}
		_, err = bot.SendVideo(vf, media.Caption)
		return err
	}
	return err
}

func (ums *UploadMediaStrategy) downloadMedia(media *core.Media) (*core.File, error) {
	mediaPath := path.Join(os.TempDir(), path.Base(media.URL))
	file, err := ums.fd.Download(media.URL, mediaPath)
	if err != nil {
		ums.l.Errorf("video download error: %v", err)
		return nil, err
	}

	ums.l.Infof("file downloaded: %s %0.2fMB", file.Name, float64(file.Size)/1024/1024)

	return file, nil
}
