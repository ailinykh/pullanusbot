package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateFfmpegConverter is a basic FfmpegConverter factory
func CreateFfmpegConverter(l core.ILogger) *FfmpegConverter {
	return &FfmpegConverter{l}
}

// FfmpegConverter implements core.IVideoConverter and core.IVideoFactory using ffmpeg
type FfmpegConverter struct {
	l core.ILogger
}

// Convert is a core.IVideoConverter interface implementation
func (c *FfmpegConverter) Convert(vf *core.Video, bitrate int) (*core.Video, error) {
	path := path.Join(os.TempDir(), vf.Name+"_converted.mp4")
	cmd := fmt.Sprintf(`ffmpeg -v error -y -i "%s" -pix_fmt yuv420p -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2" "%s"`, vf.Path, path)
	if bitrate > 0 {
		cmd = fmt.Sprintf(`ffmpeg -v error -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -v error -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, vf.Path, bitrate/1024, vf.Path, bitrate/1024, path)
	}
	c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		os.Remove(path)
		c.l.Error(err)
		return nil, errors.New(string(out))
	}

	return c.CreateVideo(path)
}

// GetCodec is a core.IVideoConverter interface implementation
func (c *FfmpegConverter) GetCodec(path string) string {
	ffprobe, err := c.getFFProbe(path)
	if err != nil {
		c.l.Error(err)
		return "unknown"
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		c.l.Error(err)
		return "unknown"
	}

	return stream.CodecName
}

// CreateMedia is a core.IMediaFactory interface implementation
func (c *FfmpegConverter) CreateMedia(url string, author *core.User) ([]*core.Media, error) {
	ffprobe, err := c.getFFProbe(url)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	if ffprobe.Format.FormatName == "image2" {
		return []*core.Media{{URL: url, Codec: stream.CodecName, Type: core.TPhoto}}, nil
	}

	return []*core.Media{{URL: url, Codec: stream.CodecName, Type: core.TVideo}}, nil
}

// CreateVideo is a core.IVideoSplitter interface implementation
func (c *FfmpegConverter) Split(video *core.Video, limit int) ([]*core.Video, error) {
	duration, n := 0, 0
	var videos = []*core.Video{}
	for duration < video.Duration {
		path := fmt.Sprintf("%s-%d.mp4", video.File.Path, n)
		cmd := fmt.Sprintf(`ffmpeg -v error -y -i %s -ss %d -fs %d %s`, video.File.Path, duration, limit, path)
		c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
		out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
		if err != nil {
			c.l.Error(err)
			return nil, errors.New(string(out))
		}

		file, err := c.CreateVideo(path)
		if err != nil {
			c.l.Error(err)
			os.Remove(path)
			if err.Error() == "file is too short" {
				// the last piece might be shorter than a second
				// example: https://youtu.be/1MLRCczBKn8
				duration += 1
				continue
			} else {
				return nil, err
			}
		}
		// defer file.Dispose()

		videos = append(videos, file)
		duration += file.Duration
		n++
	}
	return videos, nil
}

// CreateVideo is a core.IVideoFactory interface implementation
func (c *FfmpegConverter) CreateVideo(path string) (*core.Video, error) {
	c.l.Infof("create video: %s", strings.ReplaceAll(path, os.TempDir(), "$TMPDIR/"))
	ffprobe, err := c.getFFProbe(path)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	duration, err := strconv.ParseFloat(ffprobe.Format.Duration, 32)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	if duration < 2 {
		c.l.Errorf("expected duration at least 2 seconds, got %f", duration)
		return nil, errors.New("file is too short")
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	scale := "320:-1"
	if stream.Width < stream.Height {
		scale = "-1:320"
	}

	thumb, err := c.createThumb(path, scale)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	bitrate, _ := strconv.Atoi(stream.BitRate) // empty for .gif

	stat, err := os.Stat(path)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	return &core.Video{
		File:     core.File{Name: stat.Name(), Path: path, Size: stat.Size()},
		Width:    stream.Width,
		Height:   stream.Height,
		Bitrate:  bitrate,
		Duration: int(duration),
		Codec:    stream.CodecName,
		Thumb:    thumb}, nil
}

func (c *FfmpegConverter) getFFProbe(file string) (*ffpResponse, error) {
	cmd := fmt.Sprintf(`ffprobe -v error -of json -show_streams -show_format "%s"`, file)
	c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		c.l.Error(err)
		return nil, errors.New(string(out))
	}

	var resp ffpResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	return &resp, nil
}

func (c *FfmpegConverter) createThumb(videoPath string, scale string) (*core.Image, error) {
	thumbPath := videoPath + ".jpg"

	cmd := fmt.Sprintf(`ffmpeg -v error -y -i "%s" -ss 00:00:01.000 -vframes 1 -filter:v scale="%s" "%s"`, videoPath, scale, thumbPath)
	c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		c.l.Error(err)
		return nil, errors.New(string(out))
	}

	ffprobe, err := c.getFFProbe(thumbPath)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	return &core.Image{
		File:   core.File{Name: path.Base(thumbPath), Path: thumbPath},
		Width:  ffprobe.Streams[0].Width,
		Height: ffprobe.Streams[0].Height,
	}, nil
}
