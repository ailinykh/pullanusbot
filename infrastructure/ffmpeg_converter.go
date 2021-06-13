package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateFfmpegConverter() *FfmpegConverter {
	return &FfmpegConverter{}
}

type FfmpegConverter struct{}

// core.IVideoFileConverter
func (c *FfmpegConverter) Convert(vf *core.VideoFile, bitrate int) (*core.VideoFile, error) {
	convertedVideoFilePath := path.Join(os.TempDir(), vf.Name+"_converted.mp4")
	cmd := fmt.Sprintf(`ffmpeg -y -i "%s" -pix_fmt yuv420p -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2" "%s"`, vf.Path, convertedVideoFilePath)
	if bitrate > 0 {
		cmd = fmt.Sprintf(`ffmpeg -v error -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, vf.Path, bitrate/1024, vf.Path, bitrate/1024, convertedVideoFilePath)
	}
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		os.Remove(convertedVideoFilePath)
		return nil, errors.New(string(out))
	}

	return c.CreateVideoFile(convertedVideoFilePath)
}

func (c *FfmpegConverter) CreateVideoFile(path string) (*core.VideoFile, error) {
	ffprobe, err := c.getFFProbe(path)
	if err != nil {
		return nil, err
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		return nil, err
	}

	bitrate, _ := strconv.Atoi(stream.BitRate) // empty for .gif

	duration, err := strconv.ParseFloat(ffprobe.Format.Duration, 32)
	if err != nil {
		return nil, err
	}

	thumbpath := path + ".jpg"
	scale := "320:-1"
	if stream.Width < stream.Height {
		scale = "-1:320"
	}
	cmd := fmt.Sprintf(`ffmpeg -v error -i "%s" -ss 00:00:01.000 -vframes 1 -filter:v scale="%s" "%s"`, path, scale, thumbpath)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		os.Remove(thumbpath)
		return nil, errors.New(string(out))
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &core.VideoFile{
		File:      core.File{Name: stat.Name(), Path: path},
		Width:     stream.Width,
		Height:    stream.Height,
		Bitrate:   bitrate,
		Duration:  int(duration),
		Codec:     stream.CodecName,
		ThumbPath: thumbpath}, nil
}

func (c *FfmpegConverter) getFFProbe(file string) (*ffpResponse, error) {
	cmd := fmt.Sprintf(`ffprobe -v error -of json -show_streams -show_format "%s"`, file)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, errors.New(string(out))
	}

	var resp ffpResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type ffpResponse struct {
	Streams []ffpStream `json:"streams"`
	Format  ffpFormat   `json:"format"`
}

type ffpStream struct {
	Index     int    `json:"index"`
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	BitRate   string `json:"bit_rate"`
}

type ffpFormat struct {
	Filename       string `json:"filename"`
	NbStreams      int    `json:"nb_streams"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
}

func (f ffpResponse) getVideoStream() (ffpStream, error) {
	for _, stream := range f.Streams {
		if stream.CodecType == "video" {
			return stream, nil
		}
	}
	return ffpStream{}, errors.New("no video stream found")
}
