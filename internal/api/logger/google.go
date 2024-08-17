package logger

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/logger"
)

func NewGoogleLogger(ctx context.Context, workingDir string) *GoogleLogger {
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

func (l *GoogleLogger) Warn(v ...interface{}) {
	l.l.WarningDepth(1, v...)
}

func (l *GoogleLogger) Error(v ...interface{}) {
	l.l.ErrorDepth(1, v...)
}

func (l *GoogleLogger) Errorf(s string, v ...interface{}) {
	l.l.ErrorDepth(1, fmt.Sprintf(s, v...))
}

func (l *GoogleLogger) Info(v ...interface{}) {
	l.l.InfoDepth(1, v...)
}

func (l *GoogleLogger) Infof(s string, v ...interface{}) {
	l.l.InfoDepth(1, fmt.Sprintf(s, v...))
}

func (l *GoogleLogger) Warning(v ...interface{}) {
	l.l.WarningDepth(1, v...)
}

func (l *GoogleLogger) Warningf(s string, v ...interface{}) {
	l.l.WarningDepth(1, fmt.Sprintf(s, v...))
}
