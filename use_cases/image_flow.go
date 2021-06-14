package use_cases

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateImageFlow(l core.ILogger, fu core.IFileUploader) *ImageFlow {
	return &ImageFlow{l, fu}
}

type ImageFlow struct {
	l  core.ILogger
	fu core.IFileUploader
}

// IImageHandler
func (f *ImageFlow) HandleImage(image *core.Image, message *core.Message, bot core.IBot) error {
	if !message.IsPrivate {
		return nil
	}

	err := image.Download()
	if err != nil {
		return err
	}

	url, err := f.fu.Upload(&image.File)
	if err != nil {
		return err
	}

	f.l.Info(url)
	return bot.SendText(url)
}
