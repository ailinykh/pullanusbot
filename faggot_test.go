package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestFaggotStat(t *testing.T) {

	stats := FaggotStat{}

	stats.Increment("player1")
	stats.Increment("player1")
	stats.Increment("player2")

	// Test Less
	if !stats.Less(1, 0) {
		t.Error("Statistics Less function incorrect behaviour")
	}

	// Test Swap
	stats.Swap(0, 1)

	if stats.stat[0].Player != "player2" {
		t.Error("Statistics Swap function incorrect behaviour")
	}
}

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

type FakeBot struct {
	replies []string
}

func (f *FakeBot) Handle(endpoint interface{}, handler interface{}) {
}

func (f *FakeBot) Send(chat tb.Recipient, message interface{}, params ...interface{}) (*tb.Message, error) {
	if f.replies == nil {
		f.replies = []string{}
	}

	f.replies = append(f.replies, message.(string))

	return &tb.Message{}, nil
}

func (f *FakeBot) Start() {
}

func (f *FakeBot) replyText() string {
	if len(f.replies) == 0 {
		return "NO_REPLIES"
	}

	return f.replies[0]
}

func tearUp(t *testing.T) func() {
	originalrootDir := rootDir
	originalBot := bot
	rootDir = path.Join(os.TempDir(), fmt.Sprintf("pullanusbot_data_%s_%s", t.Name(), randStringRunes(4)))

	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		err = os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Can't create directory: %s", rootDir)
		}
	}

	logPath := path.Join(rootDir, "log.txt")
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	defer lf.Close()
	defer logger.Init("pullanusbot", true, true, lf).Close()

	setupDB(rootDir)

	bot = &FakeBot{}
	checkErr(err)

	// tearDown
	return func() {
		os.RemoveAll(rootDir)
		rootDir = originalrootDir
		bot = originalBot
	}
}

func TestRulesCommand(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()

	faggot.rules(&tb.Message{})

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "Правила игры") {
		t.Log(text)
		t.Error("/rules command must respond rules")
	}
}

func TestRegCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()

	// It should respond only in groups
	m := getPrivateMessage()
	faggot.reg(m)
	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "команда недоступна в личных чатах") {
		t.Log(text)
		t.Error("/reg command must respond only in groups")
	}
}

func TestRegCommandAddsPlayerInGame(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()

	// Add new player to game
	m := getGroupMessage()
	faggot.reg(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "Ты в игре") {
		t.Log(text)
		t.Error("/reg command must add player to game")
	}

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
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode)

	faggot.reg(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "Ты уже в игре") {
		t.Log(text)
		t.Error("/reg command must deny player duplicating")
	}

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
	faggot := &Faggot{}
	faggot.initialize()

	// It should respond only in groups
	m := getPrivateMessage()
	faggot.play(m)
	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "команда недоступна в личных чатах") {
		t.Log(text)
		t.Error("/play command must respond only in groups")
	}
}

func TestPlayCommandRespondsNoPlayers(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()

	m := getGroupMessage()
	faggot.play(m)
	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "Зарегистрированных в игру еще нет") {
		t.Log(text)
		t.Error("/play command must respond no players")
	}
}

func TestPlayCommandRespondsNotEnoughPlayers(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode)

	faggot.play(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "Нужно как минимум два игрока") {
		t.Log(text)
		t.Error("/play command must respond not enough players")
	}
}

func TestPlayCommandNotRespondsIfGameInProgress(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			faggot.play(m)
			wg.Done()
		}()
	}

	wg.Wait()

	if len(bot.(*FakeBot).replies) > 4 {
		t.Error("/play command must not respond if game in progress")
	}
}

func TestPlayCommandRespondsWinnerAlreadyKnown(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}
	day := time.Now().Format("2006-01-02")

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", day, m.Chat.ID, player1.ID, player1.Username)

	faggot.play(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "по результатам сегодняшнего розыгрыша") {
		t.Log(text)
		t.Error("/play command must respond winner already known")
	}
}

func TestPlayCommandLaunchGameAndRespondWinner(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	// time.Sleep(6 * time.Second)
	faggot.play(m)

	replyToCallTimes := len(bot.(*FakeBot).replies)
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
	faggot := &Faggot{}
	faggot.initialize()

	// It should respond only in groups
	m := getPrivateMessage()
	faggot.all(m)
	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "команда недоступна в личных чатах") {
		t.Log(text)
		t.Error("/all command must respond only in groups")
	}
}

func TestAllCommandNotRespondsIfNoGamesPlayedYet(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	faggot.all(m)

	if len(bot.(*FakeBot).replies) > 0 {
		t.Error("/all command must not respond if no any game results presents")
	}
}

func TestAllCommandRespondsWithAllTimeStat(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-10", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-09", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-12-31", m.Chat.ID, player1.ID, player1.Username)

	faggot.all(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "за всё время") {
		t.Log(text)
		t.Error("/all command must respond with all time statistic")
	}
}

func TestStatsCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()

	// It should respond only in groups
	m := getPrivateMessage()
	faggot.stats(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "команда недоступна в личных чатах") {
		t.Log(text)
		t.Error("/stats command must respond only in groups")
	}
}

func TestStatsCommandNotRespondingWhenNoGames(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	faggot.stats(m)

	if len(bot.(*FakeBot).replies) > 0 {
		t.Error("/stats command must not respond if no any game results presents")
	}
}

func TestStatsCommandRespondsWithCurrentYearStat(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-10", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-09", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-12-31", m.Chat.ID, player1.ID, player1.Username)

	faggot.stats(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "за текущий год") {
		t.Log(text)
		t.Error("/all command must respond with all time statistic")
	}
}

func TestMeCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()

	// It should respond only in groups
	m := getPrivateMessage()
	faggot.me(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "команда недоступна в личных чатах") {
		t.Log(text)
		t.Error("/stats command must respond only in groups")
	}
}

func TestMeCommandRespondsWithPersonalStat(t *testing.T) {
	defer tearUp(t)()
	faggot := &Faggot{}
	faggot.initialize()
	m := getGroupMessage()

	player1 := m.Sender
	player2 := tb.User{ID: 1918, FirstName: "Jozeph", LastName: "Stalin", Username: "stalin", LanguageCode: "ru"}

	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player1.ID, player1.FirstName, player1.LastName, player1.Username, player1.LanguageCode)
	db.Exec("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)", m.Chat.ID, player2.ID, player2.FirstName, player2.LastName, player2.Username, player2.LanguageCode)

	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-10", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-01-09", m.Chat.ID, player1.ID, player1.Username)
	db.Exec("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)", "2019-12-31", m.Chat.ID, player1.ID, player1.Username)

	faggot.me(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "3 раз") {
		t.Log(text)
		t.Error("/me command must respond with personal statistic for all time")
	}
}
