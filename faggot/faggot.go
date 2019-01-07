package faggot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Faggot struct for game result serialization
type faggot struct {
	Day      string `json:"day"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

func (f *faggot) mention() string {
	var str strings.Builder
	str.WriteString("@")
	str.WriteString(f.Username)
	return str.String()
}

// Setup all nesessary message handlers
func Setup(bot *tb.Bot) {
	bot.Handle("/pidorules", handle(bot, rules))
	bot.Handle("/pidoreg", handle(bot, reg))
	bot.Handle("/pidor", handle(bot, play))
	bot.Handle("/pidorall", handle(bot, all))
	bot.Handle("/pidorstats", handle(bot, stats))
	bot.Handle("/pidorme", handle(bot, me))
}

func handle(bot *tb.Bot, f func(bot *tb.Bot, m *tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		// defer func() {
		// 	// recover from panic if one occured. Set err to nil otherwise.
		// 	if r := recover(); r != nil {
		// 		bot.Send(m.Chat, r)
		// 	}
		// }()

		f(bot, m)
	}
}

func rules(bot *tb.Bot, m *tb.Message) {
	bot.Send(m.Chat, i18n("rules"), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func reg(bot *tb.Bot, m *tb.Message) {
	players := loadPlayers(m.Chat)

	for _, p := range players {
		if p.ID == m.Sender.ID {
			bot.Send(m.Chat, i18n("already_in_game"))
			return
		}
	}

	players = append(players, m.Sender)
	savePlayers(m.Chat, players)

	bot.Send(m.Chat, i18n("added_to_game"))
}

func play(bot *tb.Bot, m *tb.Message) {
	log.Println("Playing faggot of the day!")

	rand.Seed(time.Now().UTC().UnixNano())

	game := loadGame(m.Chat)
	players := loadPlayers(m.Chat)

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	if len(players) == 0 {
		log.Println("No players!")
		f := faggot{Username: m.Sender.Username}
		bot.Send(m.Chat, fmt.Sprintf(i18n("no_players"), f.mention()), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		return
	} else if len(players) == 1 {
		bot.Send(m.Chat, i18n("not_enough_players"))
		return
	}

	for _, entry := range game {
		if entry.Day == day {
			log.Println("Already known!")
			phrase := fmt.Sprintf(i18n("winner_known"), entry.Username)
			bot.Send(m.Chat, phrase, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
			return
		}
	}

	winner := players[rand.Intn(len(players))]
	log.Printf("Faggot of the day is %s!", winner.Username)

	for i := 0; i <= 3; i++ {
		template := fmt.Sprintf("faggot_game_%d_%d", i, rand.Intn(5))
		phrase := i18n(template)
		log.Printf("using template: %s", template)

		if i == 3 {
			f := faggot{Username: winner.Username}
			phrase = fmt.Sprintf(phrase, f.mention())
		}

		bot.Send(m.Chat, phrase, &tb.SendOptions{ParseMode: tb.ModeMarkdown})

		r := rand.Intn(1) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	game = append(game, faggot{day, winner.ID, winner.Username})
	saveGame(m.Chat, game)
}

func all(bot *tb.Bot, m *tb.Message) {
	s := []string{i18n("faggot_all_top"), ""}
	players := map[string]int{}
	game := loadGame(m.Chat)

	for _, entry := range game {
		players[entry.Username]++
	}

	n := 0
	for player, count := range players {
		n++
		s = append(s, fmt.Sprintf(i18n("faggot_all_entry"), n, player, count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), len(loadPlayers(m.Chat))))
	bot.Send(m.Chat, strings.Join(s, "\n"), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func stats(bot *tb.Bot, m *tb.Message) {
	s := []string{i18n("faggot_stats_top"), ""}
	players := map[string]int{}
	game := loadGame(m.Chat)
	loc, _ := time.LoadLocation("Europe/Zurich")
	year := time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, loc)

	for _, entry := range game {
		t, err := time.Parse("2006-01-02", entry.Day)

		if err != nil {
			log.Println(err)
		}

		if t.After(year) {
			players[entry.Username]++
		} else {
			log.Printf("%s is less than %s", t, year)
		}
	}

	n := 0
	for player, count := range players {
		n++
		s = append(s, fmt.Sprintf(i18n("faggot_stats_entry"), n, player, count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), len(loadPlayers(m.Chat))))
	bot.Send(m.Chat, strings.Join(s, "\n"), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func me(bot *tb.Bot, m *tb.Message) {
	game := loadGame(m.Chat)
	f := faggot{Username: m.Sender.Username}
	n := 0

	for _, entry := range game {
		if entry.UserID == m.Sender.ID {
			n++
		}
	}

	bot.Send(m.Chat, fmt.Sprintf(i18n("faggot_me"), f.mention(), n), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func loadPlayers(chat *tb.Chat) []*tb.User {
	filename := fmt.Sprintf("data/players%d.json", chat.ID)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		savePlayers(chat, []*tb.User{})
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	var players []*tb.User
	err = json.Unmarshal(data, &players)
	if err != nil {
		log.Fatal(err)
	}
	return players
}

func savePlayers(chat *tb.Chat, players []*tb.User) {
	filename := fmt.Sprintf("data/players%d.json", chat.ID)
	json, err := json.MarshalIndent(players, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filename, json, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func loadGame(chat *tb.Chat) []faggot {
	filename := fmt.Sprintf("data/game%d.json", chat.ID)

	// if err := os.Remove(filename); err != nil {
	// 	log.Fatal(err)
	// }

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		saveGame(chat, []faggot{})
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	var game []faggot
	err = json.Unmarshal(data, &game)
	if err != nil {
		log.Fatal(err)
	}
	return game
}

func saveGame(chat *tb.Chat, game []faggot) {
	log.Printf("Saving game (%d)", len(game))

	filename := fmt.Sprintf("data/game%d.json", chat.ID)
	json, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filename, json, 0644)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Game saved")
}
