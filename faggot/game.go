package faggot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var dataDir = "data"

// Game POTD structure
type Game struct {
	bot *tb.Bot
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

	return Game{bot}
}

// Start function initialize all nesessary command handlers
func (g *Game) Start() {
	g.bot.Handle("/pidorules", g.rules)
	g.bot.Handle("/pidoreg", g.reg)
	g.bot.Handle("/pidor", g.play)
	g.bot.Handle("/pidorall", g.all)
	g.bot.Handle("/pidorstats", g.stats)
	g.bot.Handle("/pidorme", g.me)

	log.Println("Game started")
}

func (g *Game) loadEntries(chatID int64) []*Entry {
	log.Printf("Loading game (%d)", chatID)
	filename := fmt.Sprintf("data/game%d.json", chatID)

	// if err := os.Remove(filename); err != nil {
	// 	log.Fatal(err)
	// }

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		g.saveEntries(chatID, []*Entry{})
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	var game []*Entry
	err = json.Unmarshal(data, &game)
	if err != nil {
		log.Fatal(err)
	}
	return game
}

func (g *Game) saveEntries(chatID int64, entries []*Entry) {
	log.Printf("Saving game (%d, %d)", chatID, len(entries))

	filename := fmt.Sprintf("data/game%d.json", chatID)
	json, err := json.MarshalIndent(entries, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filename, json, 0644)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Game saved (%d)", chatID)
}

func (g *Game) loadPlayers(chatID int64) []*Player {
	log.Printf("Loading players (%d)", chatID)

	filename := path.Join(dataDir, fmt.Sprintf("players%d.json", chatID))

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		g.savePlayers(chatID, []*Player{})
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err, chatID)
	}

	var players []*Player
	err = json.Unmarshal(data, &players)
	if err != nil {
		log.Fatal(err, chatID)
	}
	return players
}

func (g *Game) savePlayers(chatID int64, players []*Player) {
	log.Printf("Saving players (%d, %d)", chatID, len(players))

	filename := path.Join(dataDir, fmt.Sprintf("players%d.json", chatID))
	json, err := json.MarshalIndent(players, "", "  ")

	if err != nil {
		log.Fatal(err, chatID)
	}

	err = ioutil.WriteFile(filename, json, 0644)

	if err != nil {
		log.Fatal(err, chatID)
	}

	log.Printf("Players saved (%d)", chatID)
}

var replyTo = func(bot *tb.Bot, m *tb.Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (g *Game) reply(m *tb.Message, text string) {
	replyTo(g.bot, m, text)
	// g.bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (g *Game) rules(m *tb.Message) {
	g.reply(m, i18n("rules"))
}

func (g *Game) reg(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	log.Printf("Registering new player (%d)", m.Chat.ID)

	players := g.loadPlayers(m.Chat.ID)

	for _, p := range players {
		if p.ID == m.Sender.ID {
			log.Printf("Player already in game! (%d, %d)", m.Sender.ID, m.Chat.ID)
			g.reply(m, i18n("already_in_game"))
			return
		}
	}

	players = append(players, &Player{User: m.Sender})
	g.savePlayers(m.Chat.ID, players)

	log.Printf("Player added to game (%d, %d)", m.Sender.ID, m.Chat.ID)
	g.reply(m, i18n("added_to_game"))
}

func (g *Game) play(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	log.Printf("POTD: Playing pidor of the day! (%d)", m.Chat.ID)

	rand.Seed(time.Now().UTC().UnixNano())

	entries := g.loadEntries(m.Chat.ID)
	players := g.loadPlayers(m.Chat.ID)

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	if len(players) == 0 {
		log.Printf("POTD: No players! (%d)", m.Chat.ID)
		player := Player{User: m.Sender}
		g.reply(m, fmt.Sprintf(i18n("no_players"), player.mention()))
		return
	} else if len(players) == 1 {
		log.Printf("POTD: Not enough players! (%d)", m.Chat.ID)
		g.reply(m, i18n("not_enough_players"))
		return
	}

	for _, entry := range entries {
		if entry.Day == day {
			log.Printf("POTD: Already known! (%d)", m.Chat.ID)
			phrase := fmt.Sprintf(i18n("winner_known"), entry.Username)
			g.reply(m, phrase)
			return
		}
	}

	winner := players[rand.Intn(len(players))]
	log.Printf("POTD: Pidor of the day is %s! (%d)", winner.Username, m.Chat.ID)

	for i := 0; i <= 3; i++ {
		template := fmt.Sprintf("faggot_game_%d_%d", i, rand.Intn(5))
		phrase := i18n(template)
		log.Printf("POTD: using template: %s (%d)", template, m.Chat.ID)

		if i == 3 {
			phrase = fmt.Sprintf(phrase, winner.mention())
		}

		g.reply(m, phrase)

		r := rand.Intn(1) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	entries = append(entries, &Entry{day, winner.ID, winner.Username})
	g.saveEntries(m.Chat.ID, entries)
}

func (g *Game) all(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_all_top"), ""}
	players := map[string]int{}
	entries := g.loadEntries(m.Chat.ID)

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

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), len(g.loadPlayers(m.Chat.ID))))
	g.reply(m, strings.Join(s, "\n"))
}

func (g *Game) stats(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_stats_top"), ""}
	players := map[string]int{}
	entries := g.loadEntries(m.Chat.ID)
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
			log.Printf("%s is not this year!", t)
		}
	}

	n := 0
	for player, count := range players {
		n++
		s = append(s, fmt.Sprintf(i18n("faggot_stats_entry"), n, player, count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), len(g.loadPlayers(m.Chat.ID))))
	g.reply(m, strings.Join(s, "\n"))
}

func (g *Game) me(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("not_available_for_private"))
		return
	}

	game := g.loadEntries(m.Chat.ID)
	player := Player{User: m.Sender}
	n := 0

	for _, entry := range game {
		if entry.UserID == m.Sender.ID {
			n++
		}
	}

	g.reply(m, fmt.Sprintf(i18n("faggot_me"), player.mention(), n))
}
