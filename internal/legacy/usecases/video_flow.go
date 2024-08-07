package usecases

import (
	"fmt"
	"math"
	"os"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

// CreateVideoFlow is a basic VideoFlow factory
func CreateVideoFlow(l core.ILogger, videoFactory core.IVideoFactory, converter core.IVideoConverter) *VideoFlow {
	return &VideoFlow{l, converter, videoFactory}
}

// VideoFlow represents convert file to video logic
type VideoFlow struct {
	l            core.ILogger
	converter    core.IVideoConverter
	videoFactory core.IVideoFactory
}

// HandleDocument is a core.IDocumentHandler protocol implementation
func (flow *VideoFlow) HandleDocument(document *core.Document, message *core.Message, bot core.IBot) error {
	vf, err := flow.videoFactory.CreateVideo(document.File.Path)
	if err != nil {
		flow.l.Error(err)
		bot.SendText(err.Error())
		return err
	}
	defer vf.Dispose()

	expectedBitrate := int(math.Min(float64(vf.Bitrate), 568320))

	if expectedBitrate != vf.Bitrate {
		flow.l.Infof("Converting %s because of bitrate", vf.Name)
		cvf, err := flow.converter.Convert(vf, expectedBitrate)
		if err != nil {
			flow.l.Error(err)
			return err
		}
		defer cvf.Dispose()
		fi1, _ := os.Stat(vf.Path)
		fi2, _ := os.Stat(cvf.Path)
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>src: %.2f MB (%d kb/s) %s\ndst: %.2f MB (%d kb/s) %s</i>", vf.Name, message.Sender.DisplayName(), float32(fi1.Size())/1048576, vf.Bitrate/1024, vf.Codec, float32(fi2.Size())/1048576, cvf.Bitrate/1024, cvf.Codec)
		_, err = bot.SendVideo(cvf, caption)
		if err != nil {
			flow.l.Error(err)
			return err
		}
		return bot.Delete(message)
	}

	if vf.Codec != "h264" {
		flow.l.Infof("Converting %s because of codec %s", vf.Name, vf.Codec)
		cvf, err := flow.converter.Convert(vf, 0)
		if err != nil {
			flow.l.Error(err)
			return err
		}
		defer cvf.Dispose()
		fi1, _ := os.Stat(vf.Path)
		fi2, _ := os.Stat(cvf.Path)
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>src: %.2f MB (%d kb/s) %s\ndst: %.2f MB (%d kb/s) %s</i>", vf.Name, message.Sender.DisplayName(), float32(fi1.Size())/1048576, vf.Bitrate/1024, vf.Codec, float32(fi2.Size())/1048576, cvf.Bitrate/1024, cvf.Codec)
		_, err = bot.SendVideo(cvf, caption)
		if err != nil {
			flow.l.Error(err)
			return err
		}
		return bot.Delete(message)
	}

	flow.l.Infof("No need to convert %s", vf.Name)
	caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, message.Sender.DisplayName())
	_, err = bot.SendVideo(vf, caption)
	if err != nil {
		return err
	}
	return bot.Delete(message)
}
