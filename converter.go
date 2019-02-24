package main

import (
	"fmt"
	"log"
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
		defer os.Remove(sourceFile)
		defer os.Remove(destinationFile)

		log.Println("Converter: Downloading video...")

		b.Download(&m.Document.File, sourceFile)

		log.Println("Converter: Video downloaded. Converting...")

		bitrate, err := c.getBitrate(sourceFile)
		if err != nil {
			log.Printf("Converter: Bitrate converting error: %s", err)
			return
		}
		bitrate /= 1024
		if bitrate > 600 {
			bitrate = 555
		}

		cmd := exec.Command("/bin/sh", "-c", "ffmpeg -y -i \""+sourceFile+"\" -c:v libx264 -preset medium -b:v "+strconv.Itoa(bitrate)+"k -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i \""+sourceFile+"\" -c:v libx264 -preset medium -b:v "+strconv.Itoa(bitrate)+"k -pass 2 -b:a 128k \""+destinationFile+"\"")
		err = cmd.Run()
		if err != nil {
			log.Printf("Converter: Video converting error: %s", err)
			return
		}
		// cmd.Wait()
		log.Println("Converter: Video converted successfully")

		fi, _ := os.Stat(destinationFile)

		video := tb.Video{File: tb.FromDisk(destinationFile)}
		video.Caption = fmt.Sprintf("*%s* (by %s)\n_Original size: %.2f MB\nConverted size: %.2f MB_", m.Document.FileName, m.Sender.Username, float32(m.Document.FileSize)/1048576, float32(fi.Size())/1048576)
		video.SupportsStreaming = true
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

func (c *Converter) getBitrate(file string) (int, error) {
	out, err := exec.Command("/bin/sh", "-c", "ffprobe -v error -select_streams v:0 -show_entries stream=bit_rate -of default=noprint_wrappers=1:nokey=1 \""+file+"\"").Output()
	if err != nil {
		log.Printf("Converter: Bitrate retreiving error: %s", err)
		return 0, err
	}

	bitrate, err := strconv.Atoi(string(out[:len(out)-1])) // "strip EOL"
	if err != nil {
		log.Printf("Converter: Bitrate converting error: %s", err)
		return 0, err
	}

	return bitrate, nil
}
