package use_cases

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/stretchr/testify/assert"
)

func Test_RulesCommand_DeliversRules(t *testing.T) {
	game, _, l := makeSUT(LocalizerDict{"faggot_rules": "Game rules:"})
	expected := l.I18n("faggot_rules")
	rules := game.Rules()

	assert.Equal(t, rules, expected)
}

func Test_RulesCommand_DeliversRulesInDifferentTranslations(t *testing.T) {
	game, _, l := makeSUT(LocalizerDict{"faggot_rules": "Правила игры:"})
	expected := l.I18n("faggot_rules")
	rules := game.Rules()

	assert.Equal(t, rules, expected)
}

func Test_Add_ReturnsErrorOnStorageError(t *testing.T) {
	game, storage, _ := makeSUT()
	storage.err = errors.New("Unexpected error")
	player := core.Player{Username: "Faggot"}
	message := game.Add(player, storage)

	assert.Equal(t, message, storage.err.Error())
}

func Test_Add_AppendsPlayerInGameOnlyOnce(t *testing.T) {
	game, storage, localizer := makeSUT(LocalizerDict{
		"faggot_added_to_game":   "Player added",
		"faggot_already_in_game": "Player already in game",
	})
	player := core.Player{Username: "Faggot"}

	message := game.Add(player, storage)

	assert.Equal(t, storage.players, []core.Player{player})
	assert.Equal(t, message, localizer.I18n("faggot_added_to_game"))

	message = game.Add(player, storage)

	assert.Equal(t, storage.players, []core.Player{player})
	assert.Equal(t, message, localizer.I18n("faggot_already_in_game"))
}

func Test_Play_RespondsWithNoPlayers(t *testing.T) {
	game, storage, localizer := makeSUT(LocalizerDict{
		"faggot_no_players": "Nobody in game. So you win, %s!",
	})
	player := core.Player{Username: "Faggot"}
	messages := game.Play(player, storage)
	expected := []string{localizer.I18n("faggot_no_players", player.Username)}
	assert.Equal(t, messages, expected)
}

func Test_Play_RespondsNotEnoughPlayers(t *testing.T) {
	game, storage, localizer := makeSUT(LocalizerDict{
		"faggot_not_enough_players": "Not enough players",
	})
	player := core.Player{Username: "Faggot"}
	game.Add(player, storage)

	messages := game.Play(player, storage)
	expected := []string{localizer.I18n("faggot_not_enough_players")}
	assert.Equal(t, messages, expected)
}

func Test_Play_RespondsWinnerAlreadyKnown(t *testing.T) {
	game, storage, localizer := makeSUT(LocalizerDict{
		"faggot_game_0_0":     "0",
		"faggot_game_1_0":     "1",
		"faggot_game_2_0":     "2",
		"faggot_game_3_0":     "3 %s",
		"faggot_winner_known": "Winner already known %s",
	})
	player1 := core.Player{Username: "Faggot1"}
	player2 := core.Player{Username: "Faggot2"}
	game.Add(player1, storage)
	game.Add(player2, storage)

	messages := game.Play(player1, storage)
	expected := []string{"0", "1", "2", fmt.Sprintf("3 @%s", storage.rounds[0].Winner.Username)}
	assert.Equal(t, messages, expected)

	messages = game.Play(player1, storage)
	expected = []string{localizer.I18n("faggot_winner_known", storage.rounds[0].Winner.Username)}
	assert.Equal(t, messages, expected)
}

func Test_Stats_RespondsWithDescendingResultsForCurrentYear(t *testing.T) {
	year := strconv.Itoa(time.Now().Year())
	game, storage, _ := makeSUT(LocalizerDict{
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

	player1 := core.Player{Username: "Faggot1"}
	player2 := core.Player{Username: "Faggot2"}
	player3 := core.Player{Username: "Faggot3"}

	storage.rounds = []core.Round{
		{Day: year + "-01-01", Winner: player2},
		{Day: "2020-01-02", Winner: player3},
		{Day: year + "-01-02", Winner: player3},
		{Day: year + "-01-03", Winner: player3},
		{Day: year + "-01-04", Winner: player3},
		{Day: year + "-01-05", Winner: player1},
		{Day: year + "-01-06", Winner: player1},
	}

	message := game.Stats(storage)
	assert.Equal(t, strings.Split(message, "\n"), expected)
}

func Test_All_RespondsWithDescendingResultsForAllTime(t *testing.T) {
	game, storage, _ := makeSUT(LocalizerDict{
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

	player1 := core.Player{Username: "Faggot1"}
	player2 := core.Player{Username: "Faggot2"}
	player3 := core.Player{Username: "Faggot3"}

	storage.rounds = []core.Round{
		{Day: "2021-01-01", Winner: player2},
		{Day: "2020-01-02", Winner: player3},
		{Day: "2020-01-02", Winner: player3},
		{Day: "2021-01-03", Winner: player3},
		{Day: "2021-01-04", Winner: player3},
		{Day: "2021-01-05", Winner: player1},
		{Day: "2021-01-06", Winner: player1},
	}

	message := game.All(storage)
	assert.Equal(t, strings.Split(message, "\n"), expected)
}

func Test_Me_RespondsWithPersonalStat(t *testing.T) {
	game, storage, localizer := makeSUT(LocalizerDict{
		"faggot_me": "username:%s,scores:%d",
	})

	player1 := core.Player{Username: "Faggot1"}
	player2 := core.Player{Username: "Faggot2"}

	storage.rounds = []core.Round{
		{Day: "2021-01-01", Winner: player2},
		{Day: "2021-01-05", Winner: player1},
		{Day: "2021-01-06", Winner: player1},
	}

	var message string
	message = game.Me(player1, storage)
	assert.Equal(t, message, localizer.I18n("faggot_me", player1.Username, 2))

	message = game.Me(player2, storage)
	assert.Equal(t, message, localizer.I18n("faggot_me", player2.Username, 1))
}

// Helpers

func makeSUT(args ...interface{}) (*GameFlow, *GameStorageMock, *LocalizerMock) {
	dict := LocalizerDict{}
	storage := &GameStorageMock{players: []core.Player{}}

	for _, arg := range args {
		switch opt := arg.(type) {
		case LocalizerDict:
			dict = opt
		}
	}

	l := &LocalizerMock{dict: dict}
	game := &GameFlow{l}
	return game, storage, l
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
	players []core.Player
	rounds  []core.Round
	err     error
}

func (s *GameStorageMock) AddPlayer(player core.Player) error {
	if s.err != nil {
		return s.err
	}

	s.players = append(s.players, player)
	return nil
}

func (s *GameStorageMock) GetPlayers() ([]core.Player, error) {
	if s.err != nil {
		return []core.Player{}, s.err
	}

	return s.players, nil
}

func (s *GameStorageMock) AddRound(round core.Round) error {
	if s.err != nil {
		return s.err
	}

	s.rounds = append(s.rounds, round)
	return nil
}

func (s *GameStorageMock) GetRounds() ([]core.Round, error) {
	if s.err != nil {
		return []core.Round{}, s.err
	}

	return s.rounds, nil
}
