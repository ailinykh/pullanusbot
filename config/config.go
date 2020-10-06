package config

import (
	"os"
	"strconv"

	"github.com/google/logger"
)

type Config struct {
	BotToken     string
	ReportChatID int64
	Debug        bool
	SmsAPIKey    string
	VpnHost      string
	WorkingDir   string
}

func Get() *Config {
	token := os.Getenv("BOT_TOKEN")
	if len(token) == 0 {
		logger.Fatal("BOT_TOKEN required")
	}

	reportChatID, err := strconv.ParseInt(os.Getenv("ADMIN_CHAT_ID"), 10, 64)
	if err != nil {
		logger.Fatal("ADMIN_CHAT_ID required")
	}

	debug, err := strconv.ParseBool(os.Getenv("DEV"))
	if err != nil {
		debug = false
	}

	workingDir := os.Getenv("WORKING_DIR")
	if len(workingDir) == 0 {
		workingDir = "data"
	}

	c := &Config{
		BotToken:     token,
		ReportChatID: reportChatID,
		Debug:        debug,
		WorkingDir:   workingDir,
	}

	return c
}
