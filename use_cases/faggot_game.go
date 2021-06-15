package use_cases

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateGameFlow(l core.ILocalizer) GameFlow {
	return GameFlow{l}
}

type GameFlow struct {
	l core.ILocalizer
}

func (flow *GameFlow) Rules() string {
	return flow.l.I18n("faggot_rules")
}

func (flow *GameFlow) Add(player *core.User, storage core.IGameStorage) string {
	players, _ := storage.GetPlayers()
	for _, p := range players {
		if p == player {
			return flow.l.I18n("faggot_already_in_game")
		}
	}

	err := storage.AddPlayer(player)

	if err != nil {
		return "Unexpected error"
	}

	return flow.l.I18n("faggot_added_to_game")
}

func (flow *GameFlow) Play(player *core.User, storage core.IGameStorage) []string {
	players, _ := storage.GetPlayers()
	switch len(players) {
	case 0:
		return []string{flow.l.I18n("faggot_no_players", player.Username)}
	case 1:
		return []string{flow.l.I18n("faggot_not_enough_players")}
	}

	games, _ := storage.GetRounds()
	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	for _, r := range games {
		if r.Day == day {
			return []string{flow.l.I18n("faggot_winner_known", r.Winner.Username)}
		}
	}

	winner := players[rand.Intn(len(players))]
	round := &core.Round{Day: day, Winner: winner}
	storage.AddRound(round)

	phrases := make([]string, 0, 4)
	for i := 0; i <= 3; i++ {
		templates := []string{}
		for _, key := range flow.l.AllKeys() {
			if strings.HasPrefix(key, fmt.Sprintf("faggot_game_%d", i)) {
				templates = append(templates, key)
			}
		}
		template := templates[rand.Intn(len(templates))]
		phrase := flow.l.I18n(template)

		if i == 3 {
			// TODO: implementation detail leaked
			phrase = flow.l.I18n(template, "@"+winner.Username)
		}

		phrases = append(phrases, phrase)
	}

	return phrases
}

func (flow *GameFlow) All(storage core.IGameStorage) string {
	entries, _ := flow.getStat(storage)
	messages := []string{flow.l.I18n("faggot_all_top"), ""}
	for i, e := range entries {
		message := flow.l.I18n("faggot_all_entry", i+1, e.Player.Username, e.Score)
		messages = append(messages, message)
	}
	messages = append(messages, "", flow.l.I18n("faggot_all_bottom", len(entries)))
	return strings.Join(messages, "\n")
}

func (flow *GameFlow) Stats(storage core.IGameStorage) string {
	year := strconv.Itoa(time.Now().Year())
	rounds, _ := storage.GetRounds()
	entries := []Stat{}

	for _, r := range rounds {
		if strings.HasPrefix(r.Day, year) {
			index := Find(entries, r.Winner.Username)
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

	messages := []string{flow.l.I18n("faggot_stats_top"), ""}
	for i, e := range entries {
		message := flow.l.I18n("faggot_stats_entry", i+1, e.Player.Username, e.Score)
		messages = append(messages, message)
	}
	messages = append(messages, "", flow.l.I18n("faggot_stats_bottom", len(entries)))
	return strings.Join(messages, "\n")
}

func (flow *GameFlow) Me(player *core.User, storage core.IGameStorage) string {
	entries, _ := flow.getStat(storage)
	score := 0
	for _, e := range entries {
		if e.Player == player {
			score = e.Score
		}
	}
	return flow.l.I18n("faggot_me", player.Username, score)
}

func (flow *GameFlow) getStat(storage core.IGameStorage) ([]Stat, error) {
	entries := []Stat{}
	rounds, err := storage.GetRounds()

	if err != nil {
		return nil, err
	}

	for _, r := range rounds {
		index := Find(entries, r.Winner.Username)
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
