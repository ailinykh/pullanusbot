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
func (f *ImageFlow) HandleImage(file *core.File, message core.Message, bot core.IBot) error {
	if !message.IsPrivate {
		return nil
	}

	url, err := f.fu.Upload(file)
	if err != nil {
		return err
	}

	return bot.SendText(url)
}
