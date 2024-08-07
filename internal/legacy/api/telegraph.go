package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

// CreateTelegraphAPI is a default Telegraph factory
func CreateTelegraphAPI() *Telegraph {
	return &Telegraph{}
}

// Telegraph uploads files to telegra.ph
type Telegraph struct {
}

type telegraphImage struct {
	Src string `json:"src"`
}

// Upload is a core.IFileUploader interface implementation
func (t *Telegraph) Upload(file *core.File) (core.URL, error) {
	fd, err := os.Open(file.Path)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", file.Name)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, fd)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", "https://telegra.ph/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var images []telegraphImage
	err = json.Unmarshal(body2, &images)
	if err != nil {
		return "", err
	}

	url := "https://telegra.ph" + images[0].Src
	return url, nil
}
