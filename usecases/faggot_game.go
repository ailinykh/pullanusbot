package usecases

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateGameFlow is a simple GameFlow factory
func CreateGameFlow(l core.ILogger, t core.ILocalizer, s core.IGameStorage) *GameFlow {
	return &GameFlow{l, t, s}
}

// GameFlow represents faggot game logic
type GameFlow struct {
	l core.ILogger
	t core.ILocalizer
	s core.IGameStorage
}

// Rules of the game
func (flow *GameFlow) Rules(message *core.Message, bot core.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n("faggot_not_available_for_private"))
		return err
	}
	_, err := bot.SendText(flow.t.I18n("faggot_rules"))
	return err
}

// Add a new player to game
func (flow *GameFlow) Add(message *core.Message, bot core.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n("faggot_not_available_for_private"))
		return err
	}
	players, _ := flow.s.GetPlayers(message.ChatID)
	for _, p := range players {
		if p.ID == message.Sender.ID {
			if p.FirstName != message.Sender.FirstName || p.LastName != message.Sender.LastName || p.Username != message.Sender.Username {
				_ = flow.s.UpdatePlayer(message.ChatID, message.Sender)
				_, err := bot.SendText(flow.t.I18n("faggot_info_updated"))
				return err
			}
			_, err := bot.SendText(flow.t.I18n("faggot_already_in_game"))
			return err
		}
	}

	err := flow.s.AddPlayer(message.ChatID, message.Sender)
	if err != nil {
		return err
	}

	_, err = bot.SendText(flow.t.I18n("faggot_added_to_game"))
	return err
}

var mutex sync.Mutex

// Play game
func (flow *GameFlow) Play(message *core.Message, bot core.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n("faggot_not_available_for_private"))
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()

	players, _ := flow.s.GetPlayers(message.ChatID)
	switch len(players) {
	case 0:
		_, err := bot.SendText(flow.t.I18n("faggot_no_players", message.Sender.DisplayName()))
		return err
	case 1:
		_, err := bot.SendText(flow.t.I18n("faggot_not_enough_players"))
		return err
	}

	games, _ := flow.s.GetRounds(message.ChatID)
	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	for _, r := range games {
		if r.Day == day {
			_, err := bot.SendText(flow.t.I18n("faggot_winner_known", r.Winner.DisplayName()))
			return err
		}
	}

	winner := players[rand.Intn(len(players))]

	if !bot.IsUserMemberOfChat(winner, message.ChatID) {
		_, err := bot.SendText(flow.t.I18n("faggot_winner_left"))
		return err
	}

	round := &core.Round{Day: day, Winner: winner}
	flow.s.AddRound(message.ChatID, round)

	for i := 0; i <= 3; i++ {
		templates := []string{}
		for _, key := range flow.t.AllKeys() {
			if strings.HasPrefix(key, fmt.Sprintf("faggot_game_%d", i)) {
				templates = append(templates, key)
			}
		}
		template := templates[rand.Intn(len(templates))]
		phrase := flow.t.I18n(template)

		if i == 3 {
			// TODO: implementation detail leaked
			if len(winner.Username) == 0 {
				phrase = flow.t.I18n(template, fmt.Sprintf(`<a href="tg://user?id=%d">%s %s</a>`, winner.ID, winner.FirstName, winner.LastName))
			} else {
				phrase = flow.t.I18n(template, "@"+winner.Username)
			}
		}

		_, err := bot.SendText(phrase)
		if err != nil {
			//TODO: logger?
		}

		if os.Getenv("GO_ENV") != "testing" {
			r := rand.Intn(3) + 1
			time.Sleep(time.Duration(r) * time.Second)
		}
	}

	return nil
}

// All statistics for all time
func (flow *GameFlow) All(message *core.Message, bot core.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n("faggot_not_available_for_private"))
		return err
	}

	entries, _ := flow.getStat(message)
	messages := []string{flow.t.I18n("faggot_all_top"), ""}
	for i, e := range entries {
		message := flow.t.I18n("faggot_all_entry", i+1, e.Player.DisplayName(), e.Score)
		messages = append(messages, message)
	}
	messages = append(messages, "", flow.t.I18n("faggot_all_bottom", len(entries)))
	_, err := bot.SendText(strings.Join(messages, "\n"))
	return err
}

// Stats returns current year statistics
func (flow *GameFlow) Stats(message *core.Message, bot core.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n("faggot_not_available_for_private"))
		return err
	}

	year := strconv.Itoa(time.Now().Year())
	rounds, _ := flow.s.GetRounds(message.ChatID)
	entries := []Stat{}

	for _, r := range rounds {
		if strings.HasPrefix(r.Day, year) {
			index := Find(entries, r.Winner.ID)
			if index == -1 {
				entries = append(entries, Stat{Player: r.Winner, Score: 1})
			} else {
				entries[index].Score++
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].Player.Username > entries[j].Player.Username
		}
		return entries[i].Score > entries[j].Score
	})

	messages := []string{flow.t.I18n("faggot_stats_top"), ""}
	for i, e := range entries {
		message := flow.t.I18n("faggot_stats_entry", i+1, e.Player.DisplayName(), e.Score)
		messages = append(messages, message)
	}
	messages = append(messages, "", flow.t.I18n("faggot_stats_bottom", len(entries)))
	_, err := bot.SendText(strings.Join(messages, "\n"))
	return err
}

// Me returns your personal statistics
func (flow *GameFlow) Me(message *core.Message, bot core.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n("faggot_not_available_for_private"))
		return err
	}

	entries, _ := flow.getStat(message)
	score := 0
	for _, e := range entries {
		if e.Player.ID == message.Sender.ID {
			score = e.Score
		}
	}
	_, err := bot.SendText(flow.t.I18n("faggot_me", message.Sender.DisplayName(), score))
	return err
}

func (flow *GameFlow) getStat(message *core.Message) ([]Stat, error) {
	entries := []Stat{}
	rounds, err := flow.s.GetRounds(message.ChatID)

	if err != nil {
		return nil, err
	}

	for _, r := range rounds {
		index := Find(entries, r.Winner.ID)
		if index == -1 {
			entries = append(entries, Stat{Player: r.Winner, Score: 1})
		} else {
			entries[index].Score++
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	return entries, nil
}
