package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getPrivateMessage() *tb.Message {
	var m *tb.Message
	err := json.Unmarshal([]byte(`{
		"message_id": 1488,
		"from": {
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "hitler",
			"id": 1488
		},
		"chat": {
			"id": 1488,
			"type": "private",
			"title": "Private chat",
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "hitler"
		}
	}`), &m)
	if err != nil {
		panic(err)
	}
	return m
}

func getGroupMessage() *tb.Message {
	var m *tb.Message
	err := json.Unmarshal([]byte(`{
		"message_id": 1488,
		"from": {
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "hitler",
			"id": 1488
		},
		"chat": {
			"id": -1488,
			"type": "group",
			"title": "Group chat",
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "hitler"
		}
	}`), &m)
	if err != nil {
		panic(err)
	}
	return m
}

func tearUp(t *testing.T) func() {
	defaultWorkingDir := workingDir
	workingDir = path.Join(os.TempDir(), fmt.Sprintf("pullanusbot_data_%s_%s", t.Name(), randStringRunes(4)))
	setupdb(workingDir)

	// tearDown
	return func() {
		os.RemoveAll(workingDir)
		workingDir = defaultWorkingDir
		// Restore replyTo (possibly mocked by test case)
		replyTo = func(bot IBot, m *tb.Message, text string) {
			bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		}
	}
}

func TestRulesCommand(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "Правила игры") {
			t.Log(text)
			t.Error("/rules command must respond rules")
		}
	}
	faggot.rules(&tb.Message{})
}

func TestRegCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/reg command must respond only in groups")
		}
	}
	faggot.reg(m)
}

func TestRegCommandAddsPlayerInGame(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	// Add new player to game
	m := getGroupMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "Ты в игре") {
			t.Log(text)
			t.Error("/reg command must add player to game")
		}
	}
	faggot.reg(m)

	// Check player added sucessfully
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_players WHERE chat_id = ? AND user_id = ?", m.Chat.ID, m.Sender.ID).Scan(&count)
	if err == nil && count != 1 {
		t.Log(err, count)
		t.Error("Player not added to game")
	}
}

func TestRegCommandAddsEachPlayerOnlyOnce(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode)

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "Ты уже в игре") {
			t.Log(text)
			t.Error("/reg command must deny player duplicating")
		}
	}
	faggot.reg(m)

	// Check player not added twice
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_players WHERE chat_id = ? AND user_id = ?", m.Chat.ID, m.Sender.ID).Scan(&count)
	if err == nil && count != 1 {
		t.Log(err, count)
		t.Error("Player added to game twice!")
	}
}

func TestPlayCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/play command must respond only in groups")
		}
	}
	faggot.play(m)
}

func TestPlayCommandRespondsNoPlayers(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	m := getGroupMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "Зарегистрированных в игру еще нет") {
			t.Log(text)
			t.Error("/play command must respond no players")
		}
	}
	faggot.play(m)
}

func TestPlayCommandRespondsNotEnoughPlayers(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode)

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "Нужно как минимум два игрока") {
			t.Log(text)
			t.Error("/play command must respond not enough players")
		}
	}
	faggot.play(m)
}

func TestPlayCommandNotRespondsIfGameInProgress(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	var wg sync.WaitGroup
	var mutex sync.Mutex // remove it when reply to chan

	replyCount := 0
	replyTo = func(bot IBot, m *tb.Message, text string) {
		t.Log(text)
		mutex.Lock()
		replyCount++
		mutex.Unlock()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			faggot.play(m)
			wg.Done()
		}()
	}

	wg.Wait()

	if replyCount > 4 {
		t.Error("/play command must not respond if game in progress")
	}
}

func TestPlayCommandRespondsWinnerAlreadyKnown(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}
	day := time.Now().Format("2006-01-02")

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", day, m.Chat.ID, player1.ID, player1.Username)

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "по результатам сегодняшнего розыгрыша") {
			t.Log(text)
			t.Error("/play command must respond winner already known")
		}
	}
	faggot.play(m)
}

func TestPlayCommandLaunchGameAndRespondWinner(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	replyToCallTimes := 0
	replyTo = func(bot IBot, m *tb.Message, text string) {
		replyToCallTimes++
	}

	// time.Sleep(6 * time.Second)
	faggot.play(m)

	if replyToCallTimes != 4 {
		t.Errorf("/play command must respond multiple times (got %d)", replyToCallTimes)
	}

	// Check game results
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_entries WHERE chat_id = ?", m.Chat.ID).Scan(&count)
	if err == nil && count != 1 {
		t.Log(err, count)
		t.Error("/play command must play game")
	}
}

func TestAllCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/all command must respond only in groups")
		}
	}
	faggot.all(m)
}

func TestAllCommandNotRespondsIfNoGamesPlayedYet(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()
	replied := false
	replyTo = func(bot IBot, m *tb.Message, text string) {
		t.Log(text)
		replied = true
	}

	faggot.all(m)

	if replied {
		t.Error("/all command must not respond if no any game results presents")
	}
}

func TestAllCommandRespondsWithAllTimeStat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-10", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-09", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-12-31", m.Chat.ID, player1.ID, player1.Username)

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "за всё время") {
			t.Log(text)
			t.Error("/all command must respond with all time statistic")
		}
	}
	faggot.all(m)
}

func TestStatsCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/stats command must respond only in groups")
		}
	}
	faggot.stats(m)
}

func TestStatsCommandNotRespondingWhenNoGames(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	replied := false
	replyTo = func(bot IBot, m *tb.Message, text string) {
		t.Log(text)
		replied = true
	}

	faggot.stats(m)

	if replied {
		t.Error("/stats command must not respond if no any game results presents")
	}
}

func TestStatsCommandRespondsWithCurrentYearStat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-10", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-09", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-12-31", m.Chat.ID, player1.ID, player1.Username)

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "за текущий год") {
			t.Log(text)
			t.Error("/all command must respond with all time statistic")
		}
	}
	faggot.stats(m)
}

func TestMeCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/stats command must respond only in groups")
		}
	}
	faggot.me(m)
}

func TestMeCommandRespondsWithPersonalStat(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-10", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-09", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-12-31", m.Chat.ID, player1.ID, player1.Username)

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "3 раз") {
			t.Log(text)
			t.Error("/me command must respond with personal statistic for all time")
		}
	}
	faggot.me(m)
}

func TestProxyCommandRespondsWithProxyInfo(t *testing.T) {
	defer tearUp(t)()
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewFaggotGame(bot)
	m := getGroupMessage()

	replyTo = func(bot IBot, m *tb.Message, text string) {
		if !strings.Contains(text, "secret") {
			t.Log(text)
			t.Error("/proxy command must respond with proxy information")
		}
	}
	faggot.proxy(m)
}
