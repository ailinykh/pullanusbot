package main

import (
	"encoding/json"
	"os"
)

func NewServerConfigList(configPath string) ([]*serverConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var configList []*serverConfig
	err = json.Unmarshal(data, &configList)
	if err != nil {
		return nil, err
	}

	return configList, nil
}

type ServerConfig interface {
	GetKeyID() string
	GetSecretKey() string
	GetChatId() int64
	GetCommand() string
}

type serverConfig struct {
	KeyID     string  `json:"key_id"`
	SecretKey string  `json:"secret_key"`
	ChatIds   []int64 `json:"chat_ids"`
	Command   string  `json:"command"`
}

func (c *serverConfig) GetKeyID() string {
	return c.KeyID
}

func (c *serverConfig) GetSecretKey() string {
	return c.SecretKey
}

func (c *serverConfig) GetChatIds() []int64 {
	return c.ChatIds
}

func (c *serverConfig) GetCommand() string {
	return c.Command
}
