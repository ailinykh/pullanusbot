package logger

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	"github.com/google/logger"
)

func NewGoogleLogger(ctx context.Context, workingDir string) core.Logger {
	logFilePath := path.Join(workingDir, "pullanusbot.log")
	lf, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		time.Sleep(1 * time.Millisecond)
		logger.Info("closing logger")
		logger.Close()
	}()

	return &GoogleLogger{
		l: logger.Init("pullanusbot", true, false, lf),
	}
}

type GoogleLogger struct {
	l *logger.Logger
}

func (l *GoogleLogger) Debug(v ...interface{}) {
	l.l.InfoDepth(1, v...)
}

func (l *GoogleLogger) Error(v ...interface{}) {
	if len(v) > 0 {
		if s, ok := v[0].(string); ok {
			l.l.ErrorDepth(1, fmt.Sprintf(s, v[1:]...))
			return
		}
	}
	l.l.ErrorDepth(1, v...)
}

func (l *GoogleLogger) Info(v ...interface{}) {
	if len(v) > 0 {
		if s, ok := v[0].(string); ok {
			l.l.InfoDepth(1, fmt.Sprintf(s, v[1:]...))
			return
		}
	}
	l.l.InfoDepth(1, v...)
}

func (l *GoogleLogger) Warn(v ...interface{}) {
	if len(v) > 0 {
		if s, ok := v[0].(string); ok {
			l.l.WarningDepth(1, fmt.Sprintf(s, v[1:]...))
			return
		}
	}
	l.l.WarningDepth(1, v...)
}
