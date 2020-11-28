package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/google/logger"
)

func getVideo(id string) (*Video, error) {
	cmd := fmt.Sprintf(`youtube-dl -j %s`, id)
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logger.Error(cmd)
		logger.Error(err)
		return nil, err
	}

	var video Video
	err = json.Unmarshal(out, &video)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &video, nil
}

// Simple file downloader
func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
