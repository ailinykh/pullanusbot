package api

import (
	"encoding/json"
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
func (y *YoutubeAPI) CreateMedia(url string, author *core.User) ([]*core.Media, error) {
	video, err := y.getInfo(url)
	if err != nil {
		return nil, err
	}

	return []*core.Media{
		{
			URL:      video.ID,
			Caption:  video.Title,
			Duration: video.Duration,
			Type:     core.TVideo,
		},
	}, nil
}

// CreateVideo is a core.IVideoFactory interface implementation
func (y *YoutubeAPI) CreateVideo(youtubeID string) (*core.Video, error) {
	video, err := y.getInfo(youtubeID)
	if err != nil {
		return nil, err
	}

	vf, af, err := y.getFormats(video)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("youtube-%s-%s-%s.mp4", youtubeID, vf.FormatID, af.FormatID)
	path := path.Join(os.TempDir(), name)

	cmd := fmt.Sprintf("youtube-dl -f %s+%s %s -o %s", vf.FormatID, af.FormatID, youtubeID, path)
	y.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(string(out))
		return nil, err
	}

	thumb, err := y.fd.Download(video.thumb().URL) // will be disposed with Video
	if err != nil {
		return nil, err
	}

	return &core.Video{
		File:      core.File{Name: name, Path: path, Size: int64(vf.Filesize)},
		Width:     vf.Width,
		Height:    vf.Height,
		Bitrate:   0,
		Duration:  video.Duration,
		Codec:     vf.VCodec,
		ThumbPath: thumb.Path,
	}, nil
}

func (y *YoutubeAPI) getInfo(url string) (*Video, error) {
	cmd := fmt.Sprintf(`youtube-dl -j %s`, url)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, err
	}

	var video Video
	err = json.Unmarshal(out, &video)
	if err != nil {
		y.l.Error(string(out))
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
