package helpers

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateSendVideoStrategySplitDecorator(l core.ILogger, decoratee core.ISendVideoStrategy, splitter core.IVideoSplitter) core.ISendVideoStrategy {
	return &SendVideoStrategySplitDecorator{l, decoratee, splitter}
}

type SendVideoStrategySplitDecorator struct {
	l         core.ILogger
	decoratee core.ISendVideoStrategy
	splitter  core.IVideoSplitter
}

// SendMedia is a core.ISendVideoStrategy interface implementation
func (strategy *SendVideoStrategySplitDecorator) SendVideo(video *core.Video, caption string, bot core.IBot) error {
	err := strategy.decoratee.SendVideo(video, caption, bot)
	if err != nil && err.Error() == "telegram: Request Entity Too Large (400)" {
		strategy.l.Info("Fallback to splitting")
		files, err := strategy.splitter.Split(video, 50000000)
		if err != nil {
			return err
		}

		for _, file := range files {
			defer file.Dispose()
		}

		r := regexp.MustCompile(`<b>(.*)</b>`)
		match := r.FindStringSubmatch(caption)

		for i, file := range files {
			c := caption
			if len(match) > 0 {
				c = r.ReplaceAllString(caption, fmt.Sprintf("<b>[%d/%d] %s</b>", i+1, len(files), match[1]))
			}
			_, err := bot.SendVideo(file, c)
			if err != nil {
				return err
			}
		}

		strategy.l.Info("All parts successfully sent")
		return nil
	}
	return err
}
