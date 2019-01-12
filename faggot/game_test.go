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

func getGroupMessage() *Message {
	var m *Message
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

func TestRulesCommand(t *testing.T) {
	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Правила игры") {
			t.Log(text)
			t.Error("/rules command must respond rules")
		}
	}
	defer restoreReplyTo()

	faggot.rules(&Message{})
}

func TestRegCommandRespondsOnlyInGroupChat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/reg command must respond only in groups")
		}
	}
	faggot.reg(m)
}

func TestRegCommandAddsPlayerInGame(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	// Add new player to game
	m := getGroupMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Ты в игре") {
			t.Log(text)
			t.Error("/reg command must add player to game")
		}
	}
	faggot.reg(m)

	// Check player added sucessfully
	game := faggot.loadGame(m)
	for _, p := range game.Players {
		if p.ID == m.Sender.ID {
			return
		}
	}

	t.Error("Player not added to game")
}

func TestRegCommandAddsEachPlayerOnlyOnce(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  }
		],
		"entries": [
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)
	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Ты уже в игре") {
			t.Log(text)
			t.Error("/reg command must deny player duplicating")
		}
	}
	faggot.reg(m)

	// Check player not added twice
	game = faggot.loadGame(m)
	if len(game.Players) > 1 {
		t.Error("Player addet to game twice!")
	}
}

func TestPlayCommandRespondsOnlyInGroupChat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/play command must respond only in groups")
		}
	}
	faggot.play(m)
}

func TestPlayCommandRespondsNoPlayers(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	m := getGroupMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Зарегистрированных в игру еще нет") {
			t.Log(text)
			t.Error("/play command must respond no players")
		}
	}
	faggot.play(m)
}

func TestPlayCommandRespondsNotEnoughPlayers(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  }
		],
		"entries": [
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)
	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "Нужно как минимум два игрока") {
			t.Log(text)
			t.Error("/play command must respond not enough players")
		}
	}
	faggot.play(m)
}

func TestPlayCommandRespondsWinnerAlreadyKnown(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  },
		  {
			"id": 1489,
			"first_name": "Joseph",
			"last_name": "Goebbels",
			"username": "goebbels",
			"language_code": "en"
		  }
		],
		"entries": [
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")
	entry := Entry{Day: day, UserID: game.Players[1].ID, Username: game.Players[1].Username}
	game.Entries = append(game.Entries, &entry)
	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "по результатам сегодняшнего розыгрыша") {
			t.Log(text)
			t.Error("/play command must respond winner already known")
		}
	}
	faggot.play(m)
}

func TestPlayCommandLaunchGameAndRespondWinner(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  },
		  {
			"id": 1489,
			"first_name": "Joseph",
			"last_name": "Goebbels",
			"username": "goebbels",
			"language_code": "en"
		  }
		],
		"entries": [
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)

	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyToCallTimes := 0
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		replyToCallTimes++
	}

	// time.Sleep(6 * time.Second)
	faggot.play(m)

	if replyToCallTimes != 4 {
		t.Errorf("/play command must respond multiple times (got %d)", replyToCallTimes)
	}

	game = faggot.loadGame(m)

	if len(game.Entries) != 1 {
		t.Error("/play command must play game")
	}
}

func TestAllCommandRespondsOnlyInGroupChat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/all command must respond only in groups")
		}
	}
	faggot.all(m)
}

func TestAllCommandRespondsWithAllTimeStat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  },
		  {
			"id": 1489,
			"first_name": "Joseph",
			"last_name": "Goebbels",
			"username": "goebbels",
			"language_code": "en"
		  }
		],
		"entries": [
		  {
			"day": "2019-01-10",
			"user_id": 1488,
			"username": "hitler"
		  },
		  {
			"day": "2019-01-09",
			"user_id": 1488,
			"username": "hitler"
		  },
		  {
			"day": "2018-12-31",
			"user_id": 1488,
			"username": "hitler"
		  }
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)

	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "за всё время") {
			t.Log(text)
			t.Error("/all command must respond with all time statistic")
		}
	}
	faggot.all(m)
}
func TestStatsCommandRespondsOnlyInGroupChat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/stats command must respond only in groups")
		}
	}
	faggot.stats(m)
}

func TestAllCommandRespondsWithCurrentYearStat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  },
		  {
			"id": 1489,
			"first_name": "Joseph",
			"last_name": "Goebbels",
			"username": "goebbels",
			"language_code": "en"
		  }
		],
		"entries": [
		  {
			"day": "2019-01-10",
			"user_id": 1488,
			"username": "hitler"
		  },
		  {
			"day": "2019-01-09",
			"user_id": 1488,
			"username": "hitler"
		  },
		  {
			"day": "2018-12-31",
			"user_id": 1488,
			"username": "hitler"
		  }
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)

	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "за текущий год") {
			t.Log(text)
			t.Error("/all command must respond with all time statistic")
		}
	}
	faggot.stats(m)
}

func TestMeCommandRespondsOnlyInGroupChat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	// It should respond only in groups
	m := getPrivateMessage()
	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "команда недоступна в личных чатах") {
			t.Log(text)
			t.Error("/stats command must respond only in groups")
		}
	}
	faggot.me(m)
}

func TestMeCommandRespondsWithPersonalStat(t *testing.T) {
	workingDir := path.Join(os.TempDir(), fmt.Sprintf("faggot_bot_data_%s", randStringRunes(4)))
	t.Logf("Using data directory: %s", workingDir)

	bot, _ := tb.NewBot(tb.Settings{})
	faggot := NewGame(bot)
	faggot.dp = NewDataProvider(workingDir)

	defer restoreReplyTo()
	defer os.RemoveAll(workingDir)

	dataMock := []byte(`{
		"players": [
		  {
			"id": 1488,
			"first_name": "Adolf",
			"last_name": "Hitler",
			"username": "adolf",
			"language_code": "en"
		  },
		  {
			"id": 1489,
			"first_name": "Joseph",
			"last_name": "Goebbels",
			"username": "goebbels",
			"language_code": "en"
		  }
		],
		"entries": [
		  {
			"day": "2019-01-10",
			"user_id": 1488,
			"username": "hitler"
		  },
		  {
			"day": "2019-01-09",
			"user_id": 1488,
			"username": "hitler"
		  },
		  {
			"day": "2018-12-31",
			"user_id": 1488,
			"username": "hitler"
		  }
		]
	  }`)
	var game *Game
	json.Unmarshal(dataMock, &game)

	m := getGroupMessage()
	faggot.saveGame(m, game)

	replyTo = func(bot *tb.Bot, m *Message, text string) {
		if !strings.Contains(text, "3 раз") {
			t.Log(text)
			t.Error("/all command must respond with all time statistic")
		}
	}
	faggot.me(m)
}
