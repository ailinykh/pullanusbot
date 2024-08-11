package main

import "os"

func NewDefaultConfig() Config {
	return &DefaultConfig{}
}

type Config interface {
	AmqpUrl() string
	BotToken() string
	WorkingDir() string
	StringForKey(string) *string
}

type DefaultConfig struct{}

func (DefaultConfig) AmqpUrl() string {
	if amqpUrl, ok := os.LookupEnv("AMQP_URL"); ok {
		return amqpUrl
	}
	panic("Pass `AMQP_URL` via Environment variable. It shoud point to valid RabbitMQ url")
}

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

func (DefaultConfig) StringForKey(key string) *string {
	v, _ := os.LookupEnv(key)
	return &v
}
