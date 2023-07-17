package api

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateYoutubeMediaFactory(l core.ILogger, api YoutubeApi, fd core.IFileDownloader) *YoutubeMediaFactory {
	return &YoutubeMediaFactory{l, api, fd}
}

type YoutubeMediaFactory struct {
	l   core.ILogger
	api YoutubeApi
	fd  core.IFileDownloader
}

// CreateMedia is a core.IMediaFactory interface implementation
func (y *YoutubeMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	resp, err := y.api.get(url)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	video, audio, err := y.getFormats(resp)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	return []*core.Media{
		{
			URL:         video.Url,
			Title:       resp.Title,
			Description: resp.Description,
			Duration:    int(resp.Duration),
			Codec:       video.Vcodec,
			Size:        int(video.Filesize),
			Type:        core.TVideo,
		},
		{
			URL:         audio.Url,
			Title:       resp.Title,
			Description: resp.Description,
			Duration:    int(resp.Duration),
			Codec:       audio.Acodec,
			Size:        int(audio.Filesize),
			Type:        core.TAudio,
		},
	}, nil
}

func (y *YoutubeMediaFactory) getFormats(resp *YtDlpResponse) (*YtDlpFormat, *YtDlpFormat, error) {
	audio, err := resp.audioFormat()
	if err != nil {
		y.l.Error(err)
		return nil, nil, err
	}

	video, err := resp.videoFormat(50_000_000 - audio.Filesize)
	if err != nil {
		y.l.Error(err)
		return nil, nil, err
	}
	return video, audio, nil
}

// CreateVideo is a core.IVideoFactory interface implementation
func (y *YoutubeMediaFactory) CreateVideo(id string) (*core.Video, error) {
	resp, err := y.api.get(id)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	video, audio, err := y.getFormats(resp)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	name := fmt.Sprintf("youtube[%s][%s][%s].mp4", resp.Id, video.FormatNote, audio.FormatNote)
	videoPath := path.Join(os.TempDir(), name)

	cmd := fmt.Sprintf("yt-dlp -f %s+%s https://youtu.be/%s -o %s", video.FormatId, audio.FormatId, resp.Id, videoPath)
	y.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(err)
		return nil, fmt.Errorf(string(out))
	}

	thumb, err := y.getThumb(resp, video)
	if err != nil {
		return nil, err
	}

	return &core.Video{
		File:     core.File{Name: name, Path: videoPath, Size: video.Filesize + audio.Filesize},
		Width:    video.Width,
		Height:   video.Height,
		Bitrate:  0,
		Duration: int(resp.Duration),
		Codec:    video.Vcodec,
		Thumb:    thumb,
	}, nil
}

func (y *YoutubeMediaFactory) getThumb(resp *YtDlpResponse, vf *YtDlpFormat) (*core.Image, error) {
	filename := fmt.Sprintf("youtube[%s][%s].jpg", resp.Id, vf.FormatId)
	originalThumbPath := path.Join(os.TempDir(), filename+"-original")
	thumbPath := path.Join(os.TempDir(), filename)
	file, err := y.fd.Download(resp.Thumbnail, originalThumbPath)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}
	defer file.Dispose()

	scale := "320:-1"
	if vf.Width < vf.Height {
		scale = "-1:320"
	}

	cmd := fmt.Sprintf(`ffmpeg -v error -y -i "%s" -vf scale=%s "%s"`, originalThumbPath, scale, thumbPath)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(err)
		y.l.Error(string(out))
		return nil, err
	}

	stat, err := os.Stat(thumbPath)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	return &core.Image{
		File:   core.File{Name: filename, Path: thumbPath, Size: stat.Size()},
		Width:  vf.Width,
		Height: vf.Height,
	}, nil
}
