package usecases

import (
	"fmt"
	"math"
	"os"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateVideoFlow is a basic VideoFlow factory
func CreateVideoFlow(l core.ILogger, f core.IVideoFactory, c core.IVideoConverter) *VideoFlow {
	return &VideoFlow{c, f, l}
}

// VideoFlow represents convert file to video logic
type VideoFlow struct {
	c core.IVideoConverter
	f core.IVideoFactory
	l core.ILogger
}

// HandleDocument is a core.IDocumentHandler protocol implementation
func (f *VideoFlow) HandleDocument(document *core.Document, message *core.Message, bot core.IBot) error {
	vf, err := f.f.CreateVideo(document.File.Path)
	if err != nil {
		f.l.Error(err)
		bot.SendText(err.Error())
		return err
	}
	defer vf.Dispose()

	expectedBitrate := int(math.Min(float64(vf.Bitrate), 568320))

	if expectedBitrate != vf.Bitrate {
		f.l.Infof("Converting %s because of bitrate", vf.Name)
		cvf, err := f.c.Convert(vf, expectedBitrate)
		if err != nil {
			f.l.Error(err)
			return err
		}
		defer cvf.Dispose()
		fi1, _ := os.Stat(vf.Path)
		fi2, _ := os.Stat(cvf.Path)
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)</i>", vf.Name, message.Sender.Username, float32(fi1.Size())/1048576, vf.Bitrate/1024, float32(fi2.Size())/1048576, cvf.Bitrate/1024)
		_, err = bot.SendVideo(cvf, caption)
		if err != nil {
			f.l.Error(err)
			return err
		}
		return bot.Delete(message)
	}

	if vf.Codec != "h264" {
		f.l.Infof("Converting %s because of codec %s", vf.Name, vf.Codec)
		cvf, err := f.c.Convert(vf, 0)
		if err != nil {
			f.l.Error(err)
			return err
		}
		defer cvf.Dispose()
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, message.Sender.Username)
		_, err = bot.SendVideo(cvf, caption)
		if err != nil {
			f.l.Error(err)
			return err
		}
		return bot.Delete(message)
	}

	f.l.Infof("No need to convert %s", vf.Name)
	caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, message.Sender.Username)
	_, err = bot.SendVideo(vf, caption)
	if err != nil {
		return err
	}
	return bot.Delete(message)
}
