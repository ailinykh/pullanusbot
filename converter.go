package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Converter helps to post video files proper way
type Converter struct {
	mutex sync.Mutex
}

// VideoFile is a simple struct for a video file representation
type VideoFile struct {
	filepath  string
	thumbpath string
	width     int
	height    int
	size      int

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
	v.size = ffpInfo.Format.size()
	v.ffpInfo = ffpInfo
	v.videoStreamInfo = &streamInfo

	return &v, nil
}

func (v VideoFile) getFFProbeInfo(file string) (*ffpResponse, error) {
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

func (v VideoFile) getThumbnail(filepath string) (string, error) {
	ffpInfo, err := v.getFFProbeInfo(filepath)
	if err != nil {
		return "", err
	}

	videoStreamInfo, err := ffpInfo.getVideoStream()
	if err != nil {
		return "", err
	}
	thumb := filepath + ".jpg"
	cmd := fmt.Sprintf(`ffmpeg -i "%s" -ss 00:00:01.000 -vframes 1 -filter:v scale="%s" "%s"`, filepath, videoStreamInfo.scale(), thumb)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		return "", err
	}

	return thumb, nil
}

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
		return "-1:90"
	}
	return "90:-1"
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

func (c *Converter) initialize() {
	bot.Handle(tb.OnDocument, c.checkMessage)
	logger.Info("successfully initialized")
}

func (c *Converter) checkMessage(m *tb.Message) {
	if m.Document.MIME[:5] == "video" {
		// Just one video at one time pls
		c.mutex.Lock()
		defer c.mutex.Unlock()

		logger.Infof("Got video! \"%s\" of type %s from %s", m.Document.FileName, m.Document.MIME, m.Sender.Username)

		if m.Document.FileSize > 20*1024*1024 {
			logger.Errorf("File is greater than 20 MB :(%d)", m.Document.FileSize)
			return
		}

		b, ok := bot.(*tb.Bot)
		if !ok {
			logger.Error("Bot cast failed")
			return
		}

		srcPath := path.Join(os.TempDir(), m.Document.FileName)
		dstPath := path.Join(os.TempDir(), "converted_"+m.Document.FileName)
		defer os.Remove(srcPath)
		defer os.Remove(dstPath)

		logger.Info("Downloading video...")
		b.Download(&m.Document.File, srcPath)
		logger.Info("Video downloaded")

		videofile, err := NewVideoFile(srcPath)
		if err != nil {
			logger.Errorf("Can't create video file for %s, %v", srcPath, err)
			return
		}
		defer os.Remove(videofile.filepath)
		defer os.Remove(videofile.thumbpath)

		if videofile.ffpInfo.Format.NbStreams == 1 {
			logger.Error("Assuming gif file. Skipping...")
			return
		}

		srcBitrate := videofile.videoStreamInfo.bitrate()
		dstBitrate := int(math.Min(float64(srcBitrate), 568320))

		logger.Infof("Source file bitrate: %d, destination file bitrate: %d", srcBitrate, dstBitrate)

		if srcBitrate != dstBitrate {
			logger.Info("Bitrates not equal. Converting...")
			cmd := fmt.Sprintf(`ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, srcPath, dstBitrate/1024, srcPath, dstBitrate/1024, dstPath)
			err = exec.Command("/bin/sh", "-c", cmd).Run()
			if err != nil {
				logger.Errorf("Video converting error: %v", err)
				return
			}
			logger.Info("Video converted successfully")
		}

		fi, err := os.Stat(dstPath)
		if os.IsNotExist(err) {
			logger.Info("Destination file not exists. Sending original...")
			caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", m.Document.FileName, m.Sender.Username)
			uploadFile(videofile, m, caption)
		} else {
			logger.Info("Sending destination file...")
			caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)</i>", m.Document.FileName, m.Sender.Username, float32(m.Document.FileSize)/1048576, srcBitrate/1024, float32(fi.Size())/1048576, dstBitrate/1024)
			videofile.filepath = dstPath // It's ok
			uploadFile(videofile, m, caption)
		}
	} else {
		logger.Errorf("%s is not mpeg video", m.Document.MIME)
	}
}
