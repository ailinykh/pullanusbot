package use_cases

import (
	"math"

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

func (f *VideoFlow) Process(vf *core.VideoFile) (*core.VideoFile, error) {

	expectedBitrate := int(math.Min(float64(vf.Bitrate), 568320))

	if expectedBitrate != vf.Bitrate {
		f.l.Infof("Converting %s because of bitrate", vf.FileName)
		convertedFile, err := f.c.Convert(vf, expectedBitrate)
		if err != nil {
			f.l.Error(err)
			return nil, err
		}
		return convertedFile, nil
	}

	if vf.Codec != "h264" {
		f.l.Infof("Converting %s because of codec %s", vf.FileName, vf.Codec)
		convertedFile, err := f.c.Convert(vf, 0)
		if err != nil {
			f.l.Error(err)
			return nil, err
		}
		return convertedFile, nil
	}

	f.l.Infof("No need to convert %s", vf.FileName)
	return vf, nil
}
