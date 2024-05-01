package api

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

type YoutubeApi interface {
	get(string) (*YtDlpResponse, error)
}

func CreateYtDlpApi(cookie string, l core.ILogger) YoutubeApi {
	return &YtDlpApi{cookie, l}
}

type YtDlpApi struct {
	cookie string
	l      core.ILogger
}

func (api *YtDlpApi) get(url string) (*YtDlpResponse, error) {
	args := []string{
		"--quiet",
		"--no-warnings",
		"--dump-json",
	}

	if len(api.cookie) > 0 {
		args = append(args, "--cookies", api.cookie)
	}

	cmd := fmt.Sprintf(`yt-dlp %s %s`, strings.Join(args, " "), url)
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
	Id           string         `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Creator      string         `json:"creator,omitempty"` // tiktok
	Track        string         `json:"track,omitempty"`   // tiktok
	Artist       string         `json:"artist,omitempty"`  // tiktok
	Duration     float64        `json:"duration"`
	ExtractorKey string         `json:"extractor_key"`
	Thumbnail    string         `json:"thumbnail"`
	Uploader     string         `json:"uploader"`
	Url          string         `json:"url,omitempty"`
	Formats      []*YtDlpFormat `json:"formats"`
}

type YtDlpFormat struct {
	Ext        string `json:"ext"`
	FormatId   string `json:"format_id"`
	Format     string `json:"format"`
	FormatNote string `json:"format_note"`
	Filesize   int64  `json:"filesize,omitempty"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Acodec     string `json:"acodec"`
	Vcodec     string `json:"vcodec"`
	Url        string `json:"url"`
}
