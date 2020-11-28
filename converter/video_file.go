package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"pullanusbot/faggot"
	i "pullanusbot/interfaces"
	"strings"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

var uploadsInProgress = faggot.ConcurrentSlice{}

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

// UploadFinishedCallback ...
// default behaviour - remove original message
func UploadFinishedCallback(bot i.Bot, m *tb.Message) {
	err := bot.Delete(m)
	if err != nil {
		logger.Error(err)
	}
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
	cmd := fmt.Sprintf(`ffprobe -v error -of json -show_streams -show_format "%s"`, file)
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
func (v *VideoFile) Upload(bot i.Bot, m *tb.Message, caption string, cb func(i.Bot, *tb.Message)) error {
	video := tb.Video{File: tb.FromDisk(v.filepath)}
	video.Width = v.videoStreamInfo.Width
	video.Height = v.videoStreamInfo.Height
	video.Caption = caption
	video.Duration = v.ffpInfo.Format.duration()
	video.SupportsStreaming = true
	video.Thumbnail = &tb.Photo{File: tb.FromDisk(v.thumbpath)}

	logger.Infof("Uploading %dx%d %ds %.02fMB %s", video.Width, video.Height, video.Duration, float64(v.ffpInfo.Format.size())/1024/1024, strings.ReplaceAll(v.filepath, os.TempDir(), "$TMPDIR/"))

	go v.notify(m.Chat.ID)
	_, err := video.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	uploadsInProgress.Remove(m.Chat.ID)

	if err == nil {
		logger.Info("Video sent successfully")
		cb(bot, m)
	} else {
		logger.Error(err)
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

func (v *VideoFile) notify(id int64) {
	uploadsInProgress.Add(id)
	for {
		if uploadsInProgress.Index(id) == -1 {
			return
		}
		chat := &tb.Chat{ID: id}
		bot.Notify(chat, tb.UploadingVideo)
		time.Sleep(time.Duration(10) * time.Second)
	}
}
