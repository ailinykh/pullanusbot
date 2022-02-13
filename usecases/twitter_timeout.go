package usecases

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterTimeout(l core.ILogger, th ITweetHandler) *TwitterTimeout {
	return &TwitterTimeout{l, th, make(map[core.Message]core.Message)}
}

// TwitterTimeout is a decorator for TwitterFlow to handle API timeouts gracefully
type TwitterTimeout struct {
	l       core.ILogger
	th      ITweetHandler
	replies map[core.Message]core.Message
}

// HandleTweet is a ITweetHandler protocol implementation
func (tt *TwitterTimeout) HandleTweet(tweetID string, message *core.Message, bot core.IBot, deleteOriginal bool) error {
	err := tt.th.HandleTweet(tweetID, message, bot, deleteOriginal)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Rate limit exceeded") {
			timeout, err := tt.parseTimeout(err)
			if err != nil {
				return err
			}

			go func() {
				time.Sleep(time.Duration(timeout) * time.Second)
				tt.HandleTweet(tweetID, message, bot, deleteOriginal)
			}()

			minutes := timeout / 60
			seconds := timeout % 60
			reply := fmt.Sprintf("twitter api timeout %d min %d sec", minutes, seconds)
			sent, err := bot.SendText(reply, message)
			if err != nil {
				return err
			}
			tt.replies[*message] = *sent
			return nil
		}
		tt.l.Error(err)
	} else if sent, ok := tt.replies[*message]; ok {
		_ = bot.Delete(&sent)
		delete(tt.replies, *message)
	}
	return err
}

func (tt *TwitterTimeout) parseTimeout(err error) (int64, error) {
	r := regexp.MustCompile(`(\-?\d+)$`)
	match := r.FindStringSubmatch(err.Error())
	if len(match) < 2 {
		return 0, fmt.Errorf("rate limit not found")
	}

	limit, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, err
	}

	timeout := limit - time.Now().Unix()
	tt.l.Infof("Twitter api timeout %d seconds", timeout)
	timeout = int64(math.Max(float64(timeout), 2)) // Twitter api timeout might be negative
	return timeout, nil
}
