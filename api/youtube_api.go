package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateYoutubeAPI(l core.ILogger, fd core.IFileDownloader) *YoutubeAPI {
	return &YoutubeAPI{l, fd}
}

type YoutubeAPI struct {
	l  core.ILogger
	fd core.IFileDownloader
}

// CreateMedia is a core.IMediaFactory interface implementation
func (y *YoutubeAPI) CreateMedia(url string) ([]*core.Media, error) {
	video, err := y.getInfo(url)
	if err != nil {
		return nil, err
	}

	vf, _, err := y.getFormats(video)
	if err != nil {
		return nil, err
	}

	return []*core.Media{
		{
			URL:         video.ID,
			Title:       video.Title,
			Description: video.Description,
			Duration:    video.Duration,
			Codec:       vf.VCodec,
			Type:        core.TVideo,
		},
	}, nil
}

// CreateVideo is a core.IVideoFactory interface implementation
func (y *YoutubeAPI) CreateVideo(id string) (*core.Video, error) {
	video, err := y.getInfo(id)
	if err != nil {
		return nil, err
	}

	vf, af, err := y.getFormats(video)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("youtube-%s-%s-%s.mp4", id, vf.FormatID, af.FormatID)
	videoPath := path.Join(os.TempDir(), name)

	cmd := fmt.Sprintf("youtube-dl -f %s+%s https://youtu.be/%s -o %s", vf.FormatID, af.FormatID, id, videoPath)
	y.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(err)
		return nil, errors.New(string(out))
	}

	thumb := video.thumb()
	thumbPath := path.Join(os.TempDir(), name+".jpg")
	y.l.Infof("downloading thumb %s", thumb.URL)
	file, err := y.fd.Download(thumb.URL, thumbPath) // will be disposed with Video
	if err != nil {
		return nil, err
	}

	return &core.Video{
		File:     core.File{Name: name, Path: videoPath, Size: int64(vf.Filesize)},
		Width:    vf.Width,
		Height:   vf.Height,
		Bitrate:  0,
		Duration: video.Duration,
		Codec:    vf.VCodec,
		Thumb: &core.Image{
			File:   *file,
			Width:  thumb.Width,
			Height: thumb.Height,
		},
	}, nil
}

func (y *YoutubeAPI) getInfo(id string) (*Video, error) {
	cmd := fmt.Sprintf(`youtube-dl -j https://youtu.be/%s`, id) // id might start with dash, ex: -bdUoHZCf24
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(err)
		return nil, errors.New(string(out))
	}

	var video Video
	err = json.Unmarshal(out, &video)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (y *YoutubeAPI) getFormats(video *Video) (*Format, *Format, error) {
	af, err := video.audioFormat()
	if err != nil {
		return nil, nil, err
	}

	vf, err := video.formatByID("134")
	if err != nil {
		formats := video.availableFormats()
		for _, f := range formats {
			y.l.Info(f)
		}
		if len(formats) == 0 {
			return nil, nil, err
		}
		vf = formats[len(formats)-1]
	}

	return vf, af, nil
}
