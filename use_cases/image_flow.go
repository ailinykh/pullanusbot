package use_cases

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateImageFlow(l core.ILogger, fu core.IFileUploader, id core.IImageDownloader) *ImageFlow {
	return &ImageFlow{l, fu, id}
}

type ImageFlow struct {
	l  core.ILogger
	fu core.IFileUploader
	id core.IImageDownloader
}

// IImageHandler
func (f *ImageFlow) HandleImage(image *core.Image, message *core.Message, bot core.IBot) error {
	if !message.IsPrivate {
		return nil
	}

	file, err := f.id.Download(image)
	if err != nil {
		return err
	}
	//TODO: memory management
	defer file.Dispose()

	url, err := f.fu.Upload(file)
	if err != nil {
		return err
	}

	f.l.Info(url)
	_, err = bot.SendText(url)
	return err
}
