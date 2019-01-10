package faggot

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Faggot structure
type Faggot struct {
	bot *tb.Bot
	dp  *DataProvider
}

// Game struct represent game info for specific chat
type Game struct {
	Players []*Player `json:"players"`
	Entries []*Entry  `json:"entries"`
}

// Message wrapper over tb.Message
type Message struct {
	*tb.Message
	// mtx *sync.Mutex
}

// NewGame creates Game for particular bot
func NewGame(bot *tb.Bot) Faggot {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Printf("Directory not exist! Creating directory: %s", dataDir)
		err = os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Can't create directory: %s", dataDir)
		}
	}

	return Faggot{bot, &DataProvider{}}
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

func (f *Faggot) handle(fun func(*Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		// f(&Message{Message: m, mtx: &sync.Mutex{}})
		fun(&Message{Message: m})
	}
}

func (f *Faggot) loadGame(m *Message) *Game {
	log.Printf("%d LOAD: Loading game", m.Chat.ID)

	filename := fmt.Sprintf("faggot%d.json", m.Chat.ID)
	data := f.dp.loadJSON(filename)
	var game Game
	err := json.Unmarshal(data, &game)

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	log.Printf("%d LOAD: Game loaded (%d)", m.Chat.ID, len(game.Players))
	return &game
}

func (f *Faggot) saveGame(m *Message, game *Game) {
	log.Printf("%d SAVE: Saving game (%d)", m.Chat.ID, len(game.Players))

	filename := fmt.Sprintf("faggot%d.json", m.Chat.ID)
	json, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	f.dp.saveJSON(filename, json)
	log.Printf("%d SAVE: Game saved (%d)", m.Chat.ID, len(game.Players))
}

var replyTo = func(bot *tb.Bot, m *Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (f *Faggot) reply(m *Message, text string) {
	replyTo(f.bot, m, text)
	// f.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

// Print game rules
func (f *Faggot) rules(m *Message) {
	f.reply(m, i18n("rules"))
}

// Register new player
func (f *Faggot) reg(m *Message) {
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

	game.Players = append(game.Players, &Player{User: m.Sender})
	f.saveGame(m, game)

	log.Printf("%d REG:  Player added to game (%d)", m.Chat.ID, m.Sender.ID)
	f.reply(m, i18n("added_to_game"))
}

// Play POTD game
func (f *Faggot) play(m *Message) {
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
		player := Player{User: m.Sender}
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

	winner := game.Players[rand.Intn(len(game.Players))]
	log.Printf("%d POTD: Pidor of the day is %s!", m.Chat.ID, winner.Username)

	for i := 0; i <= 3; i++ {
		template := fmt.Sprintf("faggot_game_%d_%d", i, rand.Intn(5))
		phrase := i18n(template)
		log.Printf("%d POTD: using template: %s", m.Chat.ID, template)

		if i == 3 {
			phrase = fmt.Sprintf(phrase, winner.mention())
		}

		f.reply(m, phrase)

		r := rand.Intn(1) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	game.Entries = append(game.Entries, &Entry{day, winner.ID, winner.Username})
	f.saveGame(m, game)
}

// Statistics for all time
func (f *Faggot) all(m *Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_all_top"), ""}
	players := map[string]int{}
	game := f.loadGame(m)

	if len(game.Entries) == 0 {
		return
	}

	for _, entry := range game.Entries {
		players[entry.Username]++
	}

	n := 0
	for player, count := range players {
		n++
		s = append(s, fmt.Sprintf(i18n("faggot_all_entry"), n, player, count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), len(game.Players)))
	f.reply(m, strings.Join(s, "\n"))
}

// Current year statistics
func (f *Faggot) stats(m *Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_stats_top"), ""}
	players := map[string]int{}
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
			players[entry.Username]++
		} else {
			log.Printf("%d STATS: %s is not this year!", m.Chat.ID, t)
		}
	}

	n := 0
	for player, count := range players {
		n++
		s = append(s, fmt.Sprintf(i18n("faggot_stats_entry"), n, player, count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), len(game.Players)))
	f.reply(m, strings.Join(s, "\n"))
}

// Personal stat
func (f *Faggot) me(m *Message) {
	if m.Private() {
		f.reply(m, i18n("not_available_for_private"))
		return
	}

	game := f.loadGame(m)
	player := Player{User: m.Sender}
	n := 0

	for _, entry := range game.Entries {
		if entry.UserID == m.Sender.ID {
			n++
		}
	}

	f.reply(m, fmt.Sprintf(i18n("faggot_me"), player.mention(), n))
}
