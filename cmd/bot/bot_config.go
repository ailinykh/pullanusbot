package main

import (
	"encoding/json"
	"os"
)

func NewBotConfig(configPath string) (BotConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config *botConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type BotConfig interface {
	GetAmqpUrl() string
	GetBotToken() string
	GetReportChatId() *int64
	GetWorkingDir() string
}

type botConfig struct {
	AmqpUrl      string `json:"amqp_url"`
	BotToken     string `json:"bot_token"`
	ReportChatId *int64 `json:"report_chat_id,omitempty"`
	WorkingDir   string `json:"working_dir"`
}

func (c *botConfig) GetAmqpUrl() string {
	return c.AmqpUrl
}

func (c *botConfig) GetBotToken() string {
	return c.BotToken
}

func (c *botConfig) GetReportChatId() *int64 {
	return c.ReportChatId
}

func (c *botConfig) GetWorkingDir() string {
	return c.WorkingDir
}
