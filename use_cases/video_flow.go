package use_cases

import (
	"fmt"
	"math"
	"os"

	"github.com/ailinykh/pullanusbot/v2/core"
)

type VideoFlow struct {
	c core.IVideoFileConverter
	f core.IVideoFileFactory
	l core.ILogger
}

func CreateVideoFlow(l core.ILogger, f core.IVideoFileFactory, c core.IVideoFileConverter) *VideoFlow {
	return &VideoFlow{c, f, l}
}

func (f *VideoFlow) HandleDocument(document *core.Document, b core.IBot) error {
	vf, err := f.f.CreateVideoFile(document.FilePath)
	if err != nil {
		f.l.Error(err)
		return b.SendText(err.Error())
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
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)</i>", vf.Name, document.Author, float32(fi1.Size())/1048576, vf.Bitrate/1024, float32(fi2.Size())/1048576, cvf.Bitrate/1024)
		return b.SendVideoFile(cvf, caption)
	}

	if vf.Codec != "h264" {
		f.l.Infof("Converting %s because of codec %s", vf.Name, vf.Codec)
		cvf, err := f.c.Convert(vf, 0)
		if err != nil {
			f.l.Error(err)
			return err
		}
		defer cvf.Dispose()
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, document.Author)
		return b.SendVideoFile(cvf, caption)
	}

	f.l.Infof("No need to convert %s", vf.Name)
	caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, document.Author)
	return b.SendVideoFile(vf, caption)
}
