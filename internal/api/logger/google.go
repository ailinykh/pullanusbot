package logger

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/google/logger"
)

func NewGoogleLogger(ctx context.Context, workingDir string) core.ILogger {
	logFilePath := path.Join(workingDir, "pullanusbot.log")
	lf, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		time.Sleep(10 * time.Millisecond)
		logger.Info("closing logger")
		logger.Close()
	}()

	return logger.Init("pullanusbot", true, false, lf)
}
