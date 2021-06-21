package api

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"

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

	// formats := video.availableFormats()
	// for _, f := range formats {
	// 	y.l.Info(f)
	// }

	ytDlFormat := "134"
	name := "youtube-" + youtubeID + "-" + ytDlFormat + ".mp4"
	path := path.Join(os.TempDir(), name)

	cmd := fmt.Sprintf("youtube-dl -f %s+140 %s -o %s", ytDlFormat, youtubeID, path)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(string(out))
		return nil, err
	}

	thumb, err := y.fd.Download(video.thumb().URL) // will be disposed with Video
	if err != nil {
		return nil, err
	}

	format, err := video.formatByID(ytDlFormat)
	if err != nil {
		return nil, err
	}

	return &core.Video{
		File:      core.File{Name: name, Path: path, Size: int64(format.Filesize)},
		Width:     format.Width,
		Height:    format.Height,
		Bitrate:   0,
		Duration:  video.Duration,
		Codec:     format.VCodec,
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
