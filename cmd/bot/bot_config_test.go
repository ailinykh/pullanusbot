package main

import (
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_BotConfig_ReadsConfigFromFile(t *testing.T) {
	data := []byte(`
	{
		"amqp_url": "http://localhost:5672",
		"bot_token": "ABotTokenProvidedByBotFather",
		"working_dir": "."
	}
	`)
	filePath := path.Join(os.TempDir(), uuid.NewString()+".json")
	err := os.WriteFile(filePath, data, 0644)
	assert.Nil(t, err)

	config, err := NewBotConfig(filePath)

	assert.Nil(t, err)
	assert.Equal(t, config.GetAmqpUrl(), "http://localhost:5672")
	assert.Equal(t, config.GetBotToken(), "ABotTokenProvidedByBotFather")
	assert.Nil(t, config.GetReportChatId())
	assert.Equal(t, config.GetWorkingDir(), ".")
}

func Test_BotConfig_ParseOptionalReportChatId(t *testing.T) {
	data := []byte(`{ "report_chat_id": -102030405060708090 }`)
	filePath := path.Join(os.TempDir(), uuid.NewString()+".json")
	err := os.WriteFile(filePath, data, 0644)
	assert.Nil(t, err)

	config, err := NewBotConfig(filePath)

	assert.Nil(t, err)
	var expected int64 = -102030405060708090
	assert.Equal(t, *config.GetReportChatId(), expected)
}
