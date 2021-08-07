package usecases

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterTimeout(l core.ILogger, tf *TwitterFlow) *TwitterTimeout {
	return &TwitterTimeout{l, tf, make(map[core.Message]core.Message)}
}

// TwitterTimeout is a decorator for TwitterFlow to handle API timeouts gracefully
type TwitterTimeout struct {
	l       core.ILogger
	tf      *TwitterFlow
	replies map[core.Message]core.Message
}

// HandleText is a core.ITextHandler protocol implementation
func (tt *TwitterTimeout) HandleText(message *core.Message, bot core.IBot) error {
	err := tt.tf.HandleText(message, bot)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Rate limit exceeded") {
			err := tt.handleTimeout(err, message, bot)
			if strings.HasPrefix(err.Error(), "twitter api timeout") {
				sent, err := bot.SendText(err.Error(), message)
				if err != nil {
					return err
				}
				tt.replies[*message] = *sent
				return nil
			}
		}
		tt.l.Error(err)
	} else if sent, ok := tt.replies[*message]; ok {
		_ = bot.Delete(&sent)
		delete(tt.replies, *message)
	}
	return err
}

func (tt *TwitterTimeout) handleTimeout(err error, message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`(\-?\d+)$`)
	match := r.FindStringSubmatch(err.Error())
	if len(match) < 2 {
		return errors.New("rate limit not found")
	}

	limit, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return err
	}

	timeout := limit - time.Now().Unix()
	tt.l.Infof("Twitter api timeout %d seconds", timeout)
	timeout = int64(math.Max(float64(timeout), 1)) // Twitter api timeout might be negative
	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		tt.HandleText(message, bot)
	}()
	minutes := timeout / 60
	seconds := timeout % 60
	return fmt.Errorf("twitter api timeout %d min %d sec", minutes, seconds)
}
