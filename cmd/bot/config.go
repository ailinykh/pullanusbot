package main

import "os"

func NewDefaultConfig() Config {
	return &DefaultConfig{}
}

type Config interface {
	BotToken() string
	WorkingDir() string
}

type DefaultConfig struct{}

func (DefaultConfig) BotToken() string {
	if botToken, ok := os.LookupEnv("BOT_TOKEN"); ok {
		return botToken
	}
	panic("Pass `BOT_TOKEN` via Environment variable")
}

func (DefaultConfig) WorkingDir() string {
	if workingDir, ok := os.LookupEnv("WORKING_DIR"); ok {
		return workingDir
	}
	return "pullanusbot-data"
}
