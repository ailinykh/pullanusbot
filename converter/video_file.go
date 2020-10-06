package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	i "pullanusbot/interfaces"
	"strings"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// VideoFile is a simple struct for a video file representation
type VideoFile struct {
	filepath  string
	thumbpath string
	width     int
	height    int
	Size      int

	ffpInfo         *ffpResponse
	videoStreamInfo *ffpStream
}

// NewVideoFile is a simple VideoFile constructor
func NewVideoFile(path string) (*VideoFile, error) {
	v := VideoFile{}
	ffpInfo, err := v.getFFProbeInfo(path)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	streamInfo, err := ffpInfo.getVideoStream()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	thumbpath := path + ".jpg"
	cmd := fmt.Sprintf(`ffmpeg -i "%s" -ss 00:00:01.000 -vframes 1 -filter:v scale="%s" "%s"`, path, streamInfo.scale(), thumbpath)
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logger.Error(out)
		logger.Error(err)
		return nil, err
	}

	v.filepath = path
	v.thumbpath = thumbpath
	v.width = streamInfo.Width
	v.height = streamInfo.Height
	v.Size = ffpInfo.Format.size()
	v.ffpInfo = ffpInfo
	v.videoStreamInfo = &streamInfo

	return &v, nil
}

func (v *VideoFile) getFFProbeInfo(file string) (*ffpResponse, error) {
	cmd := fmt.Sprintf("ffprobe -v error -of json -show_streams -show_format \"%s\"", file)
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}

	var resp ffpResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Upload file to chat
func (v *VideoFile) Upload(bot i.Bot, m *tb.Message, caption string) error {
	video := tb.Video{File: tb.FromDisk(v.filepath)}
	video.Width = v.videoStreamInfo.Width
	video.Height = v.videoStreamInfo.Height
	video.Caption = caption
	video.Duration = v.ffpInfo.Format.duration()
	video.SupportsStreaming = true
	video.Thumbnail = &tb.Photo{File: tb.FromDisk(v.thumbpath)}

	logger.Infof("Uploading %dx%d %ds %dB %s", video.Width, video.Height, video.Duration, v.ffpInfo.Format.size(), strings.ReplaceAll(v.filepath, os.TempDir(), "$TMPDIR/"))

	bot.Notify(m.Chat, tb.UploadingVideo)
	_, err := video.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err == nil {
		logger.Info("Video sent, now removing original message")
		err = bot.Delete(m)
		if err != nil {
			logger.Error(err)
		}
	} else {
		logger.Error("Can't send video: ", err)
		bot.Send(m.Chat, fmt.Sprint(err), &tb.SendOptions{ReplyTo: m})
		return err
	}
	return nil
}

// Dispose must be invoked after video file releasing
func (v *VideoFile) Dispose() {
	os.Remove(v.filepath)
	os.Remove(v.thumbpath)
}

// Duration of video in seconds
func (v *VideoFile) Duration() int {
	return v.ffpInfo.Format.duration()
}
