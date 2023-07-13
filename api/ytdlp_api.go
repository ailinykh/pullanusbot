package api

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateYtDlpApi(l core.ILogger) *YtDlpApi {
	return &YtDlpApi{l}
}

type YtDlpApi struct {
	l core.ILogger
}

func (api *YtDlpApi) get(url string) (*YtDlpResponse, error) {
	cmd := fmt.Sprintf(`yt-dlp -j "%s"`, url)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		api.l.Error(err)
		return nil, fmt.Errorf(string(out))
	}

	var resp YtDlpResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	return &resp, nil
}

type YtDlpResponse struct {
	Id           string        `json:"id"`
	Title        string        `json:"title"`
	Creator      string        `json:"creator,omitempty"` // tiktok
	Track        string        `json:"track,omitempty"`   // tiktok
	Artist       string        `json:"artist,omitempty"`  // tiktok
	Duration     float64       `json:"duration"`
	ExtractorKey string        `json:"extractor_key"`
	Thumbnail    string        `json:"thumbnail"`
	Uploader     string        `json:"uploader"`
	Url          string        `json:"url,omitempty"`
	Formats      []YtDlpFormat `json:"formats"`
}

type YtDlpFormat struct {
	Format   string `json:"format"`
	Filesize int64  `json:"filesize"`
	Height   int64  `json:"height"`
	Width    int64  `json:"width"`
	Acodec   string `json:"acodec"`
	Vcodec   string `json:"vcodec"`
}
