package helpers_test

import (
	"errors"
	"testing"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/helpers"
	"github.com/ailinykh/pullanusbot/v2/test_helpers"
	"github.com/stretchr/testify/assert"
)

func Test_UploadMedia_DoesNotFailOnEmptyMedia(t *testing.T) {
	strategy, _, bot := makeUploadMediaStrategySUT()
	media := []*core.Media{}

	strategy.SendMedia(media, bot)

	assert.Equal(t, []string{}, bot.SentMedias)
}

func Test_UploadMedia_DoesNotFallbackOnGenericError(t *testing.T) {
	strategy, proxy, bot := makeUploadMediaStrategySUT()
	media := []*core.Media{}
	proxy.Err = errors.New("an error")

	err := strategy.SendMedia(media, bot)

	assert.Equal(t, proxy.Err, err)
}

func Test_UploadMedia_FallbackOnSpecificError(t *testing.T) {
	strategy, proxy, bot := makeUploadMediaStrategySUT()
	media := []*core.Media{{ResourceURL: "https://a-url.com"}}
	proxy.Err = errors.New("failed to get HTTP URL content")

	err := strategy.SendMedia(media, bot)

	assert.Equal(t, nil, err)
}

// Helpers
func makeUploadMediaStrategySUT() (core.ISendMediaStrategy, *test_helpers.FakeSendMediaStrategy, *test_helpers.FakeBot) {
	logger := test_helpers.CreateLogger()
	send_media_strategy := test_helpers.CreateSendMediaStrategy()
	file_downloader := test_helpers.CreateFileDownloader()
	video_factory := test_helpers.CreateVideoFactory()
	strategy := helpers.CreateUploadMediaStrategy(logger, send_media_strategy, file_downloader, video_factory)
	bot := test_helpers.CreateBot()
	return strategy, send_media_strategy, bot
}
