package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

// CreateImageFlow is a basic ImageFlow factory
func CreateImageFlow(l core.Logger, fileUploader legacy.IFileUploader, imageDownloader legacy.IImageDownloader) *ImageFlow {
	return &ImageFlow{l, fileUploader, imageDownloader}
}

// ImageFlow represents convert image to hotlink logic
type ImageFlow struct {
	l               core.Logger
	fileUploader    legacy.IFileUploader
	imageDownloader legacy.IImageDownloader
}

// HandleImage is a core.IImageHandler protocol implementation
func (flow *ImageFlow) HandleImage(image *legacy.Image, message *legacy.Message, bot legacy.IBot) error {
	if !message.IsPrivate {
		return nil
	}

	file, err := flow.imageDownloader.Download(image)
	if err != nil {
		return err
	}
	//TODO: memory management
	defer file.Dispose()

	url, err := flow.fileUploader.Upload(file)
	if err != nil {
		return err
	}

	flow.l.Info(url)
	_, err = bot.SendText(url)
	return err
}
