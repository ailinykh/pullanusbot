package api

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"gorm.io/gorm"

	tb "gopkg.in/tucnak/telebot.v2"
)

// SetupGame ...
func (t *Telebot) SetupGame(g usecases.GameFlow, conn *gorm.DB) {
	t.bot.Handle("/pidorules", func(m *tb.Message) {
		text := g.Rules()
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	t.bot.Handle("/pidoreg", func(m *tb.Message) {
		text := g.Add(makeUser(m), makeStorage(conn, m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	var mutex sync.Mutex

	t.bot.Handle("/pidor", func(m *tb.Message) {
		mutex.Lock()
		defer mutex.Unlock()

		t.logger.Infof("playing game in chat %d", m.Chat.ID)
		messages := g.Play(makeUser(m), makeStorage(conn, m, t))
		if len(messages) > 1 {
			for _, msg := range messages {
				t.bot.Send(m.Chat, msg, &tb.SendOptions{ParseMode: tb.ModeHTML})
				r := rand.Intn(3) + 1
				time.Sleep(time.Duration(r) * time.Second)
			}
		} else {
			t.bot.Send(m.Chat, messages[0], &tb.SendOptions{ParseMode: tb.ModeHTML})
		}
	})

	t.bot.Handle("/pidorall", func(m *tb.Message) {
		text := g.All(makeStorage(conn, m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	t.bot.Handle("/pidorstats", func(m *tb.Message) {
		text := g.Stats(makeStorage(conn, m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	t.bot.Handle("/pidorme", func(m *tb.Message) {
		text := g.Me(makeUser(m), makeStorage(conn, m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})
}

func makeUser(m *tb.Message) *core.User {
	return &core.User{Username: m.Sender.Username}
}

func makeStorage(conn *gorm.DB, m *tb.Message, t *Telebot) core.IGameStorage {
	storage := infrastructure.CreateGameStorage(conn, m.Chat.ID, &TelebotAdapter{m, t}, t.logger)
	return &storage
}
