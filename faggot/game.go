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

// Game POTD structure
type Game struct {
	bot *tb.Bot
	dp  *DataProvider
}

// Message wrapper over tb.Message
type Message struct {
	*tb.Message
	// mtx *sync.Mutex
}

// NewGame creates Game for particular bot
func NewGame(bot *tb.Bot) Game {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Printf("Directory not exist! Creating directory: %s", dataDir)
		err = os.MkdirAll(dataDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Can't create directory: %s", dataDir)
		}
	}

	return Game{bot, &DataProvider{}}
}

// Start function initialize all nesessary command handlers
func (g *Game) Start() {
	g.bot.Handle("/pidorules", g.handle(g.rules))
	g.bot.Handle("/pidoreg", g.handle(g.reg))
	g.bot.Handle("/pidor", g.handle(g.play))
	g.bot.Handle("/pidorall", g.handle(g.all))
	g.bot.Handle("/pidorstats", g.handle(g.stats))
	g.bot.Handle("/pidorme", g.handle(g.me))

	log.Println("Game started")
}

func (g *Game) handle(f func(*Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		// f(&Message{Message: m, mtx: &sync.Mutex{}})
		f(&Message{Message: m})
	}
}

func (g *Game) loadEntries(m *Message) []*Entry {
	log.Printf("%d LOAD: Loading game", m.Chat.ID)

	filename := fmt.Sprintf("game%d.json", m.Chat.ID)
	data := g.dp.loadJSON(filename)
	var entries []*Entry
	err := json.Unmarshal(data, &entries)

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	log.Printf("%d LOAD: Game loaded (%d)", m.Chat.ID, len(entries))
	return entries
}

func (g *Game) saveEntries(m *Message, entries []*Entry) {
	log.Printf("%d SAVE: Saving game (%d)", m.Chat.ID, len(entries))

	filename := fmt.Sprintf("game%d.json", m.Chat.ID)
	json, err := json.MarshalIndent(entries, "", "  ")

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	g.dp.saveJSON(filename, json)
	log.Printf("%d SAVE: Game saved (%d)", m.Chat.ID, len(entries))
}

func (g *Game) loadPlayers(m *Message) []*Player {
	log.Printf("%d LOAD: Loading players", m.Chat.ID)

	filename := fmt.Sprintf("players%d.json", m.Chat.ID)
	data := g.dp.loadJSON(filename)
	var players []*Player
	err := json.Unmarshal(data, &players)

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	log.Printf("%d LOAD: Players loaded (%d)", m.Chat.ID, len(players))
	return players
}

func (g *Game) savePlayers(m *Message, players []*Player) {
	log.Printf("%d SAVE: Saving players (%d)", m.Chat.ID, len(players))

	filename := fmt.Sprintf("players%d.json", m.Chat.ID)
	json, err := json.MarshalIndent(players, "", "  ")

	if err != nil {
		log.Fatal(err, m.Chat.ID)
	}

	g.dp.saveJSON(filename, json)
	log.Printf("%d SAVE: Players saved (%d)", m.Chat.ID, len(players))
}

var replyTo = func(bot *tb.Bot, m *Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (g *Game) reply(m *Message, text string) {
	replyTo(g.bot, m, text)
	// g.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

// Print game rules
func (g *Game) rules(m *Message) {
	g.reply(m, i18n("rules"))
}

// Register new player
func (g *Game) reg(m *Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	log.Printf("%d REG:  Registering new player", m.Chat.ID)

	players := g.loadPlayers(m)

	for _, p := range players {
		if p.ID == m.Sender.ID {
			log.Printf("%d REG:  Player already in game! (%d)", m.Chat.ID, m.Sender.ID)
			g.reply(m, i18n("already_in_game"))
			return
		}
	}

	players = append(players, &Player{User: m.Sender})
	g.savePlayers(m, players)

	log.Printf("%d REG:  Player added to game (%d)", m.Chat.ID, m.Sender.ID)
	g.reply(m, i18n("added_to_game"))
}

// Play POTD game
func (g *Game) play(m *Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	log.Printf("%d POTD: Playing pidor of the day!", m.Chat.ID)

	rand.Seed(time.Now().UTC().UnixNano())

	entries := g.loadEntries(m)
	players := g.loadPlayers(m)

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	if len(players) == 0 {
		log.Printf("%d POTD: No players!", m.Chat.ID)
		player := Player{User: m.Sender}
		g.reply(m, fmt.Sprintf(i18n("no_players"), player.mention()))
		return
	} else if len(players) == 1 {
		log.Printf("%d POTD: Not enough players!", m.Chat.ID)
		g.reply(m, i18n("not_enough_players"))
		return
	}

	for _, entry := range entries {
		if entry.Day == day {
			log.Printf("%d POTD: Already known!", m.Chat.ID)
			phrase := fmt.Sprintf(i18n("winner_known"), entry.Username)
			g.reply(m, phrase)
			return
		}
	}

	winner := players[rand.Intn(len(players))]
	log.Printf("%d POTD: Pidor of the day is %s!", m.Chat.ID, winner.Username)

	for i := 0; i <= 3; i++ {
		template := fmt.Sprintf("faggot_game_%d_%d", i, rand.Intn(5))
		phrase := i18n(template)
		log.Printf("%d POTD: using template: %s", m.Chat.ID, template)

		if i == 3 {
			phrase = fmt.Sprintf(phrase, winner.mention())
		}

		g.reply(m, phrase)

		r := rand.Intn(1) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	entries = append(entries, &Entry{day, winner.ID, winner.Username})
	g.saveEntries(m, entries)
}

// Statistics for all time
func (g *Game) all(m *Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_all_top"), ""}
	players := map[string]int{}
	entries := g.loadEntries(m)

	if len(entries) == 0 {
		return
	}

	for _, entry := range entries {
		players[entry.Username]++
	}

	n := 0
	for player, count := range players {
		n++
		s = append(s, fmt.Sprintf(i18n("faggot_all_entry"), n, player, count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), len(g.loadPlayers(m))))
	g.reply(m, strings.Join(s, "\n"))
}

// Current year statistics
func (g *Game) stats(m *Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_stats_top"), ""}
	players := map[string]int{}
	entries := g.loadEntries(m)
	loc, _ := time.LoadLocation("Europe/Zurich")
	currentYear := time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, loc)
	nextYear := time.Date(time.Now().Year()+1, time.January, 1, 0, 0, 0, 0, loc)

	if len(entries) == 0 {
		return
	}

	for _, entry := range entries {
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

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), len(g.loadPlayers(m))))
	g.reply(m, strings.Join(s, "\n"))
}

// Personal stat
func (g *Game) me(m *Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	game := g.loadEntries(m)
	player := Player{User: m.Sender}
	n := 0

	for _, entry := range game {
		if entry.UserID == m.Sender.ID {
			n++
		}
	}

	g.reply(m, fmt.Sprintf(i18n("faggot_me"), player.mention(), n))
}
