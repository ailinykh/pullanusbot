package converter

import (
	"errors"
	"strconv"

	"github.com/google/logger"
)

type ffpResponse struct {
	Streams []ffpStream `json:"streams"`
	Format  ffpFormat   `json:"format"`
}

func (f ffpResponse) getVideoStream() (ffpStream, error) {
	for _, stream := range f.Streams {
		if stream.CodecType == "video" {
			return stream, nil
		}
	}
	return ffpStream{}, errors.New("No video stream found")
}

type ffpStream struct {
	Index     int    `json:"index"`
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	BitRate   string `json:"bit_rate"`
}

func (s ffpStream) scale() string {
	if s.Width < s.Height {
		return "-1:320"
	}
	return "320:-1"
}

func (s ffpStream) bitrate() int {
	bitrate, _ := strconv.Atoi(s.BitRate)
	return bitrate
}

type ffpFormat struct {
	Filename       string `json:"filename"`
	NbStreams      int    `json:"nb_streams"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
}

func (f ffpFormat) duration() int {
	d, err := strconv.ParseFloat(f.Duration, 32)
	if err != nil {
		logger.Errorf("Duration error: %v (%s)", err, f.Duration)
	}
	return int(d)
}

func (f ffpFormat) size() int {
	d, err := strconv.Atoi(f.Size)
	if err != nil {
		logger.Errorf("Size error: %v (%s)", err, f.Size)
	}
	return d
}
