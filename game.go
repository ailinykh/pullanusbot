package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// IBot is a generic interface for testing
type IBot interface {
	Handle(interface{}, interface{})
	Send(tb.Recipient, interface{}, ...interface{}) (*tb.Message, error)
}

// Faggot structure
type Faggot struct {
	bot IBot
	dp  *DataProvider
}

// FaggotGame struct represent game info for specific chat
type FaggotGame struct {
	Players []*FaggotPlayer `json:"players"`
	Entries []*FaggotEntry  `json:"entries"`
}

// NewFaggotGame creates FaggotGame for particular bot
func NewFaggotGame(bot IBot) Faggot {
	dp, _ := NewDataProvider()
	return Faggot{bot: bot, dp: dp}
}

// Start function initialize all nesessary command handlers
func (f *Faggot) Start() {
	f.bot.Handle("/pidorules", f.handle(f.rules))
	f.bot.Handle("/pidoreg", f.handle(f.reg))
	f.bot.Handle("/pidor", f.handle(f.play))
	f.bot.Handle("/pidorall", f.handle(f.all))
	f.bot.Handle("/pidorstats", f.handle(f.stats))
	f.bot.Handle("/pidorme", f.handle(f.me))

	log.Println("Game started")
}

func (f *Faggot) handle(fun func(*tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		fun(m)
	}
}

func (f *Faggot) loadGame(m *tb.Message) *FaggotGame {
	log.Printf("%d LOAD: Loading game", m.Chat.ID)

	filename := fmt.Sprintf("faggot%d.json", m.Chat.ID)
	data, err := f.dp.loadJSON(filename)

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	var game FaggotGame
	err = json.Unmarshal(data, &game)

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	log.Printf("%d LOAD: Game loaded (%d)", m.Chat.ID, len(game.Players))
	return &game
}

func (f *Faggot) saveGame(m *tb.Message, game *FaggotGame) {
	log.Printf("%d SAVE: Saving game (%d)", m.Chat.ID, len(game.Players))

	filename := fmt.Sprintf("faggot%d.json", m.Chat.ID)
	json, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	success := f.dp.saveJSON(filename, json)
	log.Printf("%d SAVE: Game saved (%d) %v", m.Chat.ID, len(game.Players), success)
}

var replyTo = func(bot IBot, m *tb.Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (f *Faggot) reply(m *tb.Message, text string) {
	replyTo(f.bot, m, text)
	// f.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

// Print game rules
func (f *Faggot) rules(m *tb.Message) {
	f.reply(m, i18n("rules"))
}

// Register new player
func (f *Faggot) reg(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	log.Printf("%d REG:  Registering new player", m.Chat.ID)

	game := f.loadGame(m)

	for _, p := range game.Players {
		if p.ID == m.Sender.ID {
			log.Printf("%d REG:  Player already in game! (%d)", m.Chat.ID, m.Sender.ID)
			f.reply(m, i18n("already_in_game"))
			return
		}
	}

	game.Players = append(game.Players, &FaggotPlayer{User: m.Sender})
	f.saveGame(m, game)

	log.Printf("%d REG:  Player added to game (%d)", m.Chat.ID, m.Sender.ID)
	f.reply(m, i18n("added_to_game"))
}

// Active games sync
var activeGames = ConcurrentSlice{}

// Play POTD game
func (f *Faggot) play(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	log.Printf("%d POTD: Playing pidor of the day!", m.Chat.ID)

	rand.Seed(time.Now().UTC().UnixNano())

	game := f.loadGame(m)

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	if len(game.Players) == 0 {
		log.Printf("%d POTD: No players!", m.Chat.ID)
		player := FaggotPlayer{User: m.Sender}
		f.reply(m, fmt.Sprintf(i18n("no_players"), player.mention()))
		return
	} else if len(game.Players) == 1 {
		log.Printf("%d POTD: Not enough players!", m.Chat.ID)
		f.reply(m, i18n("not_enough_players"))
		return
	}

	for _, entry := range game.Entries {
		if entry.Day == day {
			log.Printf("%d POTD: Already known!", m.Chat.ID)
			phrase := fmt.Sprintf(i18n("winner_known"), entry.Username)
			f.reply(m, phrase)
			return
		}
	}

	if activeGames.Index(m.Chat.ID) > -1 {
		log.Printf("%d PODT: Game in progress! Do nothing!", m.Chat.ID)
		return
	}

	activeGames.Add(m.Chat.ID)

	winner := game.Players[rand.Intn(len(game.Players))]
	log.Printf("%d POTD: Pidor of the day is %s!", m.Chat.ID, winner.Username)

	for i := 0; i <= 3; i++ {
		templates := []string{}
		for key := range ru {
			if strings.HasPrefix(key, fmt.Sprintf("faggot_game_%d", i)) {
				templates = append(templates, key)
			}
		}
		template := templates[rand.Intn(len(templates))]
		phrase := i18n(template)
		log.Printf("%d POTD: using template: %s", m.Chat.ID, template)

		if i == 3 {
			phrase = fmt.Sprintf(phrase, winner.mention())
		}

		f.reply(m, phrase)

		r := rand.Intn(1) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	game.Entries = append(game.Entries, &FaggotEntry{day, winner.ID, winner.Username})
	f.saveGame(m, game)

	activeGames.Remove(m.Chat.ID)
}

// Statistics for all time
func (f *Faggot) all(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_all_top"), ""}
	stats := FaggotStat{}
	game := f.loadGame(m)

	if len(game.Entries) == 0 {
		return
	}

	for _, entry := range game.Entries {
		stats.Increment(entry.Username)
	}

	sort.Sort(sort.Reverse(stats))
	for i, stat := range stats.stat {
		s = append(s, fmt.Sprintf(i18n("faggot_all_entry"), i+1, stat.Player, stat.Count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), len(game.Players)))
	f.reply(m, strings.Join(s, "\n"))
}

// Current year statistics
func (f *Faggot) stats(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_stats_top"), ""}
	stats := FaggotStat{}
	game := f.loadGame(m)
	loc, _ := time.LoadLocation("Europe/Zurich")
	currentYear := time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, loc)
	nextYear := time.Date(time.Now().Year()+1, time.January, 1, 0, 0, 0, 0, loc)

	if len(game.Entries) == 0 {
		return
	}

	for _, entry := range game.Entries {
		t, err := time.Parse("2006-01-02", entry.Day)

		if err != nil {
			log.Println(err)
		}

		if t.After(currentYear) && t.Before(nextYear) {
			stats.Increment(entry.Username)
		}
	}

	sort.Sort(sort.Reverse(stats))
	for i, stat := range stats.stat {
		s = append(s, fmt.Sprintf(i18n("faggot_stats_entry"), i+1, stat.Player, stat.Count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), len(game.Players)))
	f.reply(m, strings.Join(s, "\n"))
}

// Personal stat
func (f *Faggot) me(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	game := f.loadGame(m)
	player := FaggotPlayer{User: m.Sender}
	n := 0

	for _, entry := range game.Entries {
		if entry.UserID == m.Sender.ID {
			n++
		}
	}

	f.reply(m, fmt.Sprintf(i18n("faggot_me"), player.mention(), n))
}
