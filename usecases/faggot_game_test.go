package usecases

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/stretchr/testify/assert"
)

func Test_RulesCommand_DeliversRules(t *testing.T) {
	game, bot, _ := makeSUT(LocalizerDict{"faggot_rules": "Game rules:"})
	message := makeMessage(1, "Faggot")

	game.Rules(message, bot)

	assert.Equal(t, bot.messages[0], "Game rules:")
}

func Test_Add_AppendsPlayerInGameOnlyOnce(t *testing.T) {
	game, bot, storage := makeSUT(LocalizerDict{
		"faggot_added_to_game":   "Player added",
		"faggot_already_in_game": "Player already in game",
	})
	message := makeMessage(1, "Faggot")

	game.Add(message, bot)

	assert.Equal(t, storage.players, []*core.User{message.Sender})
	assert.Equal(t, bot.messages[0], "Player added")

	game.Add(message, bot)

	assert.Equal(t, storage.players, []*core.User{message.Sender})
	assert.Equal(t, bot.messages[1], "Player already in game")
}

func Test_Play_RespondsWithNoPlayers(t *testing.T) {
	game, bot, _ := makeSUT(LocalizerDict{
		"faggot_no_players": "Nobody in game. So you win, %s!",
	})
	message := makeMessage(1, "Faggot")

	game.Play(message, bot)

	assert.Equal(t, bot.messages[0], "Nobody in game. So you win, Faggot!")
}

func Test_Play_RespondsNotEnoughPlayers(t *testing.T) {
	game, bot, _ := makeSUT(LocalizerDict{
		"faggot_not_enough_players": "Not enough players",
	})
	message := makeMessage(1, "Faggot")

	game.Add(message, bot)
	game.Play(message, bot)

	assert.Equal(t, bot.messages[1], "Not enough players")
}

func Test_Play_RespondsWinnerAlreadyKnown(t *testing.T) {
	game, bot, storage := makeSUT(LocalizerDict{
		"faggot_game_0_0":     "0",
		"faggot_game_1_0":     "1",
		"faggot_game_2_0":     "2",
		"faggot_game_3_0":     "3 %s",
		"faggot_winner_known": "Winner already known %s",
	})
	m1 := makeMessage(1, "Faggot1")
	m2 := makeMessage(2, "Faggot2")

	game.Add(m1, bot)
	game.Add(m2, bot)
	game.Play(m1, bot)

	winner := storage.rounds[0].Winner.Username
	assert.Equal(t, bot.messages[2], "0")
	assert.Equal(t, bot.messages[3], "1")
	assert.Equal(t, bot.messages[4], "2")
	assert.Equal(t, bot.messages[5], fmt.Sprintf("3 @%s", winner))

	game.Play(m1, bot)

	assert.Equal(t, bot.messages[6], fmt.Sprintf("Winner already known %s", winner))
}

func Test_Stats_RespondsWithDescendingResultsForCurrentYear(t *testing.T) {
	year := strconv.Itoa(time.Now().Year())
	game, bot, storage := makeSUT(LocalizerDict{
		"faggot_stats_top":    "top",
		"faggot_stats_entry":  "index:%d,player:%s,scores:%d",
		"faggot_stats_bottom": "total_players:%d",
	})

	expected := []string{
		"top",
		"",
		"index:1,player:Faggot3,scores:3",
		"index:2,player:Faggot1,scores:2",
		"index:3,player:Faggot2,scores:1",
		"",
		"total_players:3",
	}

	m1 := makeMessage(1, "Faggot1")
	m2 := makeMessage(2, "Faggot2")
	m3 := makeMessage(3, "Faggot3")

	storage.rounds = []*core.Round{
		{Day: year + "-01-01", Winner: m2.Sender},
		{Day: "2020-01-02", Winner: m3.Sender},
		{Day: year + "-01-02", Winner: m3.Sender},
		{Day: year + "-01-03", Winner: m3.Sender},
		{Day: year + "-01-04", Winner: m3.Sender},
		{Day: year + "-01-05", Winner: m1.Sender},
		{Day: year + "-01-06", Winner: m1.Sender},
	}

	game.Stats(m1, bot)
	assert.Equal(t, strings.Split(bot.messages[0], "\n"), expected)
}

func Test_All_RespondsWithDescendingResultsForAllTime(t *testing.T) {
	game, bot, storage := makeSUT(LocalizerDict{
		"faggot_all_top":    "top",
		"faggot_all_entry":  "index:%d,player:%s,scores:%d",
		"faggot_all_bottom": "total_players:%d",
	})

	expected := []string{
		"top",
		"",
		"index:1,player:Faggot3,scores:4",
		"index:2,player:Faggot1,scores:2",
		"index:3,player:Faggot2,scores:1",
		"",
		"total_players:3",
	}

	m1 := makeMessage(1, "Faggot1")
	m2 := makeMessage(2, "Faggot2")
	m3 := makeMessage(3, "Faggot3")

	storage.rounds = []*core.Round{
		{Day: "2021-01-01", Winner: m2.Sender},
		{Day: "2020-01-02", Winner: m3.Sender},
		{Day: "2020-01-02", Winner: m3.Sender},
		{Day: "2021-01-03", Winner: m3.Sender},
		{Day: "2021-01-04", Winner: m3.Sender},
		{Day: "2021-01-05", Winner: m1.Sender},
		{Day: "2021-01-06", Winner: m1.Sender},
	}

	game.All(m1, bot)
	assert.Equal(t, strings.Split(bot.messages[0], "\n"), expected)
}

func Test_Me_RespondsWithPersonalStat(t *testing.T) {
	game, bot, storage := makeSUT(LocalizerDict{
		"faggot_me": "username:%s,scores:%d",
	})

	m1 := makeMessage(1, "Faggot1")
	m2 := makeMessage(2, "Faggot2")

	storage.rounds = []*core.Round{
		{Day: "2021-01-01", Winner: m2.Sender},
		{Day: "2021-01-05", Winner: m1.Sender},
		{Day: "2021-01-06", Winner: m1.Sender},
	}

	game.Me(m1, bot)
	assert.Equal(t, bot.messages[0], fmt.Sprintf("username:%s,scores:%d", m1.Sender.Username, 2))

	game.Me(m2, bot)
	assert.Equal(t, bot.messages[1], fmt.Sprintf("username:%s,scores:%d", m2.Sender.Username, 1))
}

// Helpers

func makeMessage(id int, username string) *core.Message {
	player := &core.User{
		ID:        id,
		FirstName: "FirstName" + fmt.Sprint(id),
		LastName:  "LastName" + fmt.Sprint(id),
		Username:  username,
	}
	return &core.Message{ID: 0, Sender: player}
}

func makeSUT(args ...interface{}) (*GameFlow, *BotMock, *GameStorageMock) {
	dict := LocalizerDict{}
	storage := &GameStorageMock{players: []*core.User{}}
	bot := &BotMock{}

	for _, arg := range args {
		switch opt := arg.(type) {
		case LocalizerDict:
			dict = opt
		}
	}

	l := &LocalizerMock{dict: dict}
	game := &GameFlow{l, storage}
	return game, bot, storage
}

// LocalizerMock

type LocalizerMock struct {
	dict LocalizerDict
}

type LocalizerDict = map[string]string

func (l *LocalizerMock) I18n(key string, args ...interface{}) string {
	if val, ok := l.dict[key]; ok {
		return fmt.Sprintf(val, args...)
	}
	return key
}

func (l *LocalizerMock) AllKeys() []string {
	keys := make([]string, 0, len(l.dict))
	for k := range l.dict {
		keys = append(keys, k)
	}
	return keys
}

// GameStorageMock

type GameStorageMock struct {
	players []*core.User
	rounds  []*core.Round
}

func (s *GameStorageMock) AddPlayer(gameID int64, player *core.User) error {
	s.players = append(s.players, player)
	return nil
}

func (s *GameStorageMock) GetPlayers(gameID int64) ([]*core.User, error) {
	return s.players, nil
}

func (s *GameStorageMock) AddRound(gameID int64, round *core.Round) error {
	s.rounds = append(s.rounds, round)
	return nil
}

func (s *GameStorageMock) GetRounds(gameID int64) ([]*core.Round, error) {
	return s.rounds, nil
}

type BotMock struct {
	messages []string
}

func (BotMock) Delete(*core.Message) error                                   { return nil }
func (BotMock) SendImage(*core.Image) (*core.Message, error)                 { return nil, nil }
func (BotMock) SendAlbum([]*core.Image) ([]*core.Message, error)             { return nil, nil }
func (BotMock) SendMedia(*core.Media) (*core.Message, error)                 { return nil, nil }
func (BotMock) SendPhotoAlbum([]*core.Media) ([]*core.Message, error)        { return nil, nil }
func (BotMock) SendVideoFile(*core.VideoFile, string) (*core.Message, error) { return nil, nil }

func (b *BotMock) SendText(text string, args ...interface{}) (*core.Message, error) {
	b.messages = append(b.messages, text)
	return nil, nil
}
