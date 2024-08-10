package main

import "os"

func NewDefaultConfig() Config {
	return &DefaultConfig{}
}

type Config interface {
	AmqpUrl() string
	BotToken() string
	WorkingDir() string
	LightsailCredentials() (*string, *string)
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

func (DefaultConfig) LightsailCredentials() (*string, *string) {
	if accessKeyId, ok := os.LookupEnv("LIGHTSAIL_ACCESS_KEY_ID"); ok {
		if secretAccessKey, ok := os.LookupEnv("LIGHTSAIL_SECRET_ACCESS_KEY"); ok {
			return &accessKeyId, &secretAccessKey
		}
	}
	return nil, nil
}
