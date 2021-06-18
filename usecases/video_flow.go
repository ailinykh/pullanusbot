package usecases

import (
	"fmt"
	"math"
	"os"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateVideoFlow is a basic VideoFlow factory
func CreateVideoFlow(l core.ILogger, f core.IVideoFileFactory, c core.IVideoFileConverter) *VideoFlow {
	return &VideoFlow{c, f, l}
}

// VideoFlow represents convert file to video logic
type VideoFlow struct {
	c core.IVideoFileConverter
	f core.IVideoFileFactory
	l core.ILogger
}

// HandleDocument is a core.IDocumentHandler protocol implementation
func (f *VideoFlow) HandleDocument(document *core.Document, b core.IBot) error {
	vf, err := f.f.CreateVideoFile(document.FilePath)
	if err != nil {
		f.l.Error(err)
		b.SendText(err.Error())
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
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)</i>", vf.Name, document.Author, float32(fi1.Size())/1048576, vf.Bitrate/1024, float32(fi2.Size())/1048576, cvf.Bitrate/1024)
		_, err = b.SendVideoFile(cvf, caption)
		return err
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
		_, err = b.SendVideoFile(cvf, caption)
		return err
	}

	f.l.Infof("No need to convert %s", vf.Name)
	caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, document.Author)
	_, err = b.SendVideoFile(vf, caption)
	return err
}
