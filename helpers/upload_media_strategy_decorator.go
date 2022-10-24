package helpers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateUploadMediaDecorator(l core.ILogger, decoratee core.ISendMediaStrategy, fileDownloader core.IFileDownloader, videoFactory core.IVideoFactory) core.ISendMediaStrategy {
	return &UploadMediaDecorator{l, decoratee, fileDownloader, videoFactory}
}

type UploadMediaDecorator struct {
	l              core.ILogger
	decoratee      core.ISendMediaStrategy
	fileDownloader core.IFileDownloader
	videoFactory   core.IVideoFactory
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (decorator *UploadMediaDecorator) SendMedia(media []*core.Media, bot core.IBot) error {
	err := decorator.decoratee.SendMedia(media, bot)
	if err != nil {
		decorator.l.Error(err)
		if strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") {
			return decorator.fallbackToUploading(media[0], bot)
		}
	}

	return err
}

func (decorator *UploadMediaDecorator) fallbackToUploading(media *core.Media, bot core.IBot) error {
	if media.Size/1024/1024 > 50 {
		return fmt.Errorf("file size limit exceeded")
	}

	decorator.l.Info("send by uploading")
	file, err := decorator.downloadMedia(media)
	if err != nil {
		return err
	}
	defer file.Dispose()

	switch media.Type {
	case core.TText:
		decorator.l.Warning("unexpected media type")
	case core.TPhoto:
		image := &core.Image{File: *file}
		_, err = bot.SendImage(image, media.Caption)
		return err
	case core.TVideo:
		vf, err := decorator.videoFactory.CreateVideo(file.Path)
		if err != nil {
			decorator.l.Errorf("can't create video file for %s, %v", file.Path, err)
			return err
		}
		_, err = bot.SendVideo(vf, media.Caption)
		return err
	}
	return err
}

func (decorator *UploadMediaDecorator) downloadMedia(media *core.Media) (*core.File, error) {
	//TODO: duplicated code
	filename := path.Base(media.ResourceURL)
	if strings.Contains(filename, "?") {
		parts := strings.Split(media.ResourceURL, "?")
		filename = path.Base(parts[0])
	}

	if !strings.HasSuffix(filename, ".mp4") {
		filename = filename + ".mp4"
	}

	mediaPath := path.Join(os.TempDir(), filename)
	file, err := decorator.fileDownloader.Download(media.ResourceURL, mediaPath)
	if err != nil {
		decorator.l.Errorf("video download error: %v", err)
		return nil, err
	}

	decorator.l.Infof("file downloaded: %s %0.2fMB", file.Name, float64(file.Size)/1024/1024)

	return file, nil
}
