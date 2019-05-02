package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"

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
		log.Printf("Converter: Duration error: %s (%s)", err, f.Duration)
	}
	return int(d)
}

func (c *Converter) initialize() {
	bot.Handle(tb.OnDocument, c.checkMessage)
	log.Println("Converter: successfully initialized")
}

func (c *Converter) checkMessage(m *tb.Message) {
	if m.Document.MIME[:5] == "video" {
		// Just one video at one time pls
		c.mutex.Lock()
		defer c.mutex.Unlock()

		log.Printf("Converter: Got video! \"%s\" of type %s from %s", m.Document.FileName, m.Document.MIME, m.Sender.Username)

		if m.Document.FileSize > 20*1024*1024 {
			log.Printf("Converter: File is greater than 20 MB :(%d)", m.Document.FileSize)
			return
		}

		b, ok := bot.(*tb.Bot)
		if !ok {
			log.Println("Converter: Bot cast failed")
			return
		}

		sourceFile := path.Join(os.TempDir(), m.Document.FileName)
		destinationFile := path.Join(os.TempDir(), "converted_"+m.Document.FileName)
		destinationThumbFile := path.Join(os.TempDir(), "converted_"+m.Document.FileName+"_thumb.jpg")
		defer os.Remove(sourceFile)
		defer os.Remove(destinationFile)
		defer os.Remove(destinationThumbFile)

		log.Println("Converter: Downloading video...")
		b.Download(&m.Document.File, sourceFile)
		log.Println("Converter: Video downloaded")

		ffpInfo, err := c.getFFProbeInfo(sourceFile)
		if err != nil {
			log.Printf("Converter: FFProbe info retreiving error: %s", err)
			return
		}

		if ffpInfo.Format.NbStreams == 1 {
			log.Println("Converter: Assuming gif file. Skipping...")
			return
		}

		sourceStreamInfo, err := ffpInfo.getVideoStream()
		if err != nil {
			log.Printf("Converter: %s", err)
			return
		}

		sourceBitrate := sourceStreamInfo.bitrate()
		destinationBitrate := int(math.Min(float64(sourceBitrate), 568320))

		log.Printf("Converter: Source file bitrate: %d, destination file bitrate: %d", sourceBitrate, destinationBitrate)

		if sourceBitrate != destinationBitrate {
			log.Println("Converter: Bitrates not equal. Converting...")

			// cmd := exec.Command("/bin/sh", "-c", "ffmpeg -y -i \""+sourceFile+"\" -c:v libx264 -preset medium -b:v "+strconv.Itoa(destinationBitrate/1024)+"k -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i \""+sourceFile+"\" -c:v libx264 -preset medium -b:v "+strconv.Itoa(destinationBitrate/1024)+"k -pass 2 -b:a 128k \""+destinationFile+"\"")
			cmd := fmt.Sprintf(`ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, sourceFile, destinationBitrate/1024, sourceFile, destinationBitrate/1024, destinationFile)
			err = exec.Command("/bin/sh", "-c", cmd).Run()
			if err != nil {
				log.Printf("Converter: Video converting error: %s", err)
				return
			}
			// cmd.Wait()
			log.Println("Converter: Video converted successfully")
		}

		fi, err := os.Stat(destinationFile)
		var video tb.Video

		if os.IsNotExist(err) {
			log.Println("Converter: Destination file not exists. Sending original...")
			video = tb.Video{File: tb.FromDisk(sourceFile)}
			video.Caption = fmt.Sprintf("*%s* _(by %s)_", m.Document.FileName, m.Sender.Username)
		} else {
			log.Println("Converter: Sending destination file...")
			video = tb.Video{File: tb.FromDisk(destinationFile)}
			video.Caption = fmt.Sprintf("*%s* _(by %s)_\n_Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)_", m.Document.FileName, m.Sender.Username, float32(m.Document.FileSize)/1048576, sourceBitrate/1024, float32(fi.Size())/1048576, destinationBitrate/1024)
		}

		video.Width = sourceStreamInfo.Width
		video.Height = sourceStreamInfo.Height
		video.Duration = ffpInfo.Format.duration()
		video.SupportsStreaming = true

		// Getting thumbnail
		cmd := fmt.Sprintf(`ffmpeg -i "%s" -ss 00:00:01.000 -vframes 1 -filter:v scale="%s" "%s"`, sourceFile, sourceStreamInfo.scale(), destinationThumbFile)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			log.Printf("Converter: Thumbnail error: %s", err)
		} else {
			thumb := tb.Photo{File: tb.FromDisk(destinationThumbFile)}
			video.Thumbnail = &thumb
		}

		log.Printf("Converter: Sending file: w:%d h:%d duration:%d", video.Width, video.Height, video.Duration)

		_, err = video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		// _, err := bot.Send(m.Chat, video)
		if err == nil {
			log.Println("Converter: Video sent. Deleting original")
			err = b.Delete(m)
			if err != nil {
				log.Printf("Converter: Can't delete original message: %s", err)
			}
		} else {
			log.Printf("Converter: Can't send video: %s", err)
		}
	} else {
		log.Printf("Converter: %s is not mpeg video", m.Document.MIME)
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
