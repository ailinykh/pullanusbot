package api

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	"github.com/ailinykh/pullanusbot/v2/use_cases"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (t *Telebot) SetupGame(g use_cases.GameFlow) {
	t.bot.Handle("/pidorules", func(m *tb.Message) {
		text := g.Rules()
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	t.bot.Handle("/pidoreg", func(m *tb.Message) {
		text := g.Add(makeUser(m), makeStorage(m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	var mutex sync.Mutex

	t.bot.Handle("/pidor", func(m *tb.Message) {
		mutex.Lock()
		defer mutex.Unlock()

		messages := g.Play(makeUser(m), makeStorage(m, t))
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
		text := g.All(makeStorage(m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	t.bot.Handle("/pidorstats", func(m *tb.Message) {
		text := g.Stats(makeStorage(m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})

	t.bot.Handle("/pidorme", func(m *tb.Message) {
		text := g.Me(makeUser(m), makeStorage(m, t))
		t.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
	})
}

func makeUser(m *tb.Message) *core.User {
	return &core.User{Username: m.Sender.Username}
}

func makeStorage(m *tb.Message, t *Telebot) core.IGameStorage {
	storage := infrastructure.CreateGameStorage(m.Chat.ID, &TelebotAdapter{m, t})
	return &storage
}
