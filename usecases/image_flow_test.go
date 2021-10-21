package usecases_test

import (
	"testing"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/test_helpers"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleImage_DownloadsAndUploadsImage(t *testing.T) {
	logger := test_helpers.CreateLogger()
	file_uploader := test_helpers.CreateFileUploader()
	image_downloader := test_helpers.CreateImageDownloader()
	image_flow := usecases.CreateImageFlow(logger, file_uploader, image_downloader)

	url := "http://an-image-url.com"
	path := "/an/image/path.jpg"
	image := &core.Image{FileURL: url, File: core.File{Path: path}}

	message := &core.Message{IsPrivate: true}
	bot := test_helpers.CreateBot()

	image_flow.HandleImage(image, message, bot)

	assert.Equal(t, []string{url}, image_downloader.Downloaded)
	assert.Equal(t, []string{path}, file_uploader.Uploaded)
}
