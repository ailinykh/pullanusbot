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
}

func (f ffpFormat) duration() int {
	d, err := strconv.ParseFloat(f.Duration, 32)
	if err != nil {
		logger.Errorf("Duration error: %v (%s)", err, f.Duration)
	}
	return int(d)
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

		sourceFile := path.Join(os.TempDir(), m.Document.FileName)
		destinationFile := path.Join(os.TempDir(), "converted_"+m.Document.FileName)
		defer os.Remove(sourceFile)
		defer os.Remove(destinationFile)

		logger.Info("Downloading video...")
		b.Download(&m.Document.File, sourceFile)
		logger.Info("Video downloaded")

		ffpInfo, err := c.getFFProbeInfo(sourceFile)
		if err != nil {
			logger.Errorf("FFProbe info retreiving error: %v", err)
			return
		}

		if ffpInfo.Format.NbStreams == 1 {
			logger.Error("Assuming gif file. Skipping...")
			return
		}

		sourceStreamInfo, err := ffpInfo.getVideoStream()
		if err != nil {
			logger.Errorf("%v", err)
			return
		}

		sourceBitrate := sourceStreamInfo.bitrate()
		destinationBitrate := int(math.Min(float64(sourceBitrate), 568320))

		logger.Infof("Source file bitrate: %d, destination file bitrate: %d", sourceBitrate, destinationBitrate)

		if sourceBitrate != destinationBitrate {
			logger.Info("Bitrates not equal. Converting...")

			// cmd := exec.Command("/bin/sh", "-c", "ffmpeg -y -i \""+sourceFile+"\" -c:v libx264 -preset medium -b:v "+strconv.Itoa(destinationBitrate/1024)+"k -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i \""+sourceFile+"\" -c:v libx264 -preset medium -b:v "+strconv.Itoa(destinationBitrate/1024)+"k -pass 2 -b:a 128k \""+destinationFile+"\"")
			cmd := fmt.Sprintf(`ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, sourceFile, destinationBitrate/1024, sourceFile, destinationBitrate/1024, destinationFile)
			err = exec.Command("/bin/sh", "-c", cmd).Run()
			if err != nil {
				logger.Errorf("Video converting error: %v", err)
				return
			}
			// cmd.Wait()
			logger.Info("Video converted successfully")
		}

		fi, err := os.Stat(destinationFile)
		var video tb.Video

		if os.IsNotExist(err) {
			logger.Info("Destination file not exists. Sending original...")
			video = tb.Video{File: tb.FromDisk(sourceFile)}
			video.Caption = fmt.Sprintf("*%s* _(by %s)_", m.Document.FileName, m.Sender.Username)
		} else {
			logger.Info("Sending destination file...")
			video = tb.Video{File: tb.FromDisk(destinationFile)}
			video.Caption = fmt.Sprintf("*%s* _(by %s)_\n_Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)_", m.Document.FileName, m.Sender.Username, float32(m.Document.FileSize)/1048576, sourceBitrate/1024, float32(fi.Size())/1048576, destinationBitrate/1024)
		}

		video.Width = sourceStreamInfo.Width
		video.Height = sourceStreamInfo.Height
		video.Duration = ffpInfo.Format.duration()
		video.SupportsStreaming = true

		// Getting thumbnail
		thumb, err := c.getThumbnail(destinationFile)
		if err != nil {
			logger.Errorf("Thumbnail error: %v", err)
		} else {
			video.Thumbnail = &tb.Photo{File: tb.FromDisk(thumb)}
			defer os.Remove(thumb)
		}

		logger.Infof("Sending file: w:%d h:%d duration:%d", video.Width, video.Height, video.Duration)

		_, err = video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		// _, err := bot.Send(m.Chat, video)
		if err == nil {
			logger.Info("Video sent. Deleting original")
			err = b.Delete(m)
			if err != nil {
				logger.Errorf("Can't delete original message: %v", err)
			}
		} else {
			logger.Errorf("Can't send video: %v", err)
		}
	} else {
		logger.Errorf("%s is not mpeg video", m.Document.MIME)
	}
}

func (c *Converter) getFFProbeInfo(file string) (*ffpResponse, error) {
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

func (c *Converter) getThumbnail(filepath string) (string, error) {
	ffpInfo, err := c.getFFProbeInfo(filepath)
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
