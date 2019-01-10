package faggot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func restoreReplyTo() {
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
	}
}

func getPrivateMessage() *Message {
	var m *Message
	err := json.Unmarshal([]byte(`{
		"message_id": 1488,
		"from": {
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "ahitler",
			"id": 1488
		},
		"chat": {
			"id": 1488,
			"type": "private",
			"title": "Private chat",
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "ahitler"
		}
	}`), &m)
	if err != nil {
		panic(err)
	}
	return m
}

func getGroupMessage() *Message {
	var m *Message
	err := json.Unmarshal([]byte(`{
		"message_id": 1488,
		"from": {
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "ahitler",
			"id": 1488
		},
		"chat": {
			"id": 1488,
			"type": "group",
			"title": "Group chat",
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "ahitler"
		}
	}`), &m)
	if err != nil {
		panic(err)
	}
	return m
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestRulesCommang(t *testing.T) {
	bot, _ := tb.NewBot(tb.Settings{})
	game := NewGame(bot)
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Правила игры") {
			t.Errorf("Text must contain rules but got: %s", text)
		}
	}
	defer restoreReplyTo()

	game.rules(&Message{})
}

func TestRegCommang(t *testing.T) {
	dataDir = path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", dataDir)

	bot, _ := tb.NewBot(tb.Settings{})
	game := NewGame(bot)

	defer restoreReplyTo()
	defer func() {
		os.RemoveAll(dataDir)
	}()

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Error("/reg command must respond only in groups")
		}
	}
	game.reg(m)

	m = getGroupMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Ты в игре") {
			t.Error("/reg command must add player to game")
		}
	}
	game.reg(m)

	m = getGroupMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Ты уже в игре") {
			t.Error("/reg command must deny player duplicating")
		}
	}
	game.reg(m)
}
