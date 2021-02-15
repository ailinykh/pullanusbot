package usecases

import "github.com/ailinykh/pullanusbot/v2/core"

// CreateImageFlow is a basic ImageFlow factory
func CreateImageFlow(l core.ILogger, fileUploader core.IFileUploader, imageDownloader core.IImageDownloader) *ImageFlow {
	return &ImageFlow{l, fileUploader, imageDownloader}
}

// ImageFlow represents convert image to hotlink logic
type ImageFlow struct {
	l               core.ILogger
	fileUploader    core.IFileUploader
	imageDownloader core.IImageDownloader
}

// HandleImage is a core.IImageHandler protocol implementation
func (flow *ImageFlow) HandleImage(image *core.Image, message *core.Message, bot core.IBot) error {
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
