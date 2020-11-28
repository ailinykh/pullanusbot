package faggot

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	i "pullanusbot/interfaces"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot   i.Bot
	db    *gorm.DB
	games = ConcurrentSlice{}
)

// Game is a faggot game logic
type Game struct {
}

// Setup all the nessesary bot command handlers
func (g *Game) Setup(b i.Bot, conn *gorm.DB) {
	bot, db = b, conn
	db.AutoMigrate(&Entry{}, &Player{})

	bot.Handle("/pidorules", g.rules)
	bot.Handle("/pidoreg", g.reg)
	bot.Handle("/pidor", g.play)
	bot.Handle("/pidorall", g.all)
	bot.Handle("/pidorstats", g.stats)
	bot.Handle("/pidorme", g.me)

	logger.Info("Game started")
}

func (g *Game) reply(m *tb.Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
}

// Print game rules
func (g *Game) rules(m *tb.Message) {
	g.reply(m, i18n("faggot_rules"))
}

// Register new player
func (g *Game) reg(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	logger.Infof("%d Registering new player", m.Chat.ID)

	var count int64
	db.First(&Player{}, "chat_id = ? AND user_id = ?", m.Chat.ID, m.Sender.ID).Count(&count)

	if count > 0 {
		logger.Warningf("%d Player already in game %d", m.Chat.ID, m.Sender.ID)
		g.reply(m, i18n("faggot_already_in_game"))
		return
	}

	player := Player{ChatID: m.Chat.ID, UserID: m.Sender.ID, FirstName: m.Sender.FirstName, LastName: m.Sender.LastName, Username: m.Sender.Username, LanguageCode: m.Sender.LanguageCode}
	db.Create(&player)

	logger.Infof("%d Player added to game (%d)", m.Chat.ID, m.Sender.ID)
	g.reply(m, i18n("faggot_added_to_game"))
}

// Play POTD game
func (g *Game) play(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	logger.Infof("%d Playing pidor of the day for %s!", m.Chat.ID, day)

	var players []Player
	db.Where("chat_id = ?", m.Chat.ID).Find(&players)

	logger.Infof("%d Players found: %d", m.Chat.ID, len(players))

	switch len(players) {
	case 0:
		logger.Infof("%d No players!", m.Chat.ID)
		winner := Player{m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode}
		g.reply(m, fmt.Sprintf(i18n("faggot_no_players"), winner.mention()))
		return
	case 1:
		logger.Infof("%d Not enough players!", m.Chat.ID)
		g.reply(m, i18n("faggot_not_enough_players"))
		return
	default:
	}

	var count int64
	var entry Entry
	db.First(&entry, "chat_id = ? AND day = ?", m.Chat.ID, day).Count(&count)

	if count > 0 {
		logger.Infof("%d Already known!", m.Chat.ID)
		phrase := fmt.Sprintf(i18n("faggot_winner_known"), entry.Username)
		g.reply(m, phrase)
		return
	}

	if games.Index(m.Chat.ID) > -1 {
		logger.Infof("%d PODT: Game in progress! Do nothing!", m.Chat.ID)
		return
	}

	games.Add(m.Chat.ID)
	defer games.Remove(m.Chat.ID)

	winner := players[rand.Intn(len(players))]
	logger.Infof("%d Pidor of the day is %s!", m.Chat.ID, winner.mention())

	member, err := bot.ChatMemberOf(m.Chat, &tb.User{ID: winner.UserID, FirstName: winner.FirstName, LastName: winner.LastName, Username: winner.Username, LanguageCode: winner.LanguageCode})
	if member == nil ||
		member.Role == tb.Left ||
		member.Role == tb.Kicked {
		logger.Errorf("%d %v", m.Chat.ID, member)
		logger.Errorf("%d %v", m.Chat.ID, err)
		g.reply(m, i18n("faggot_winner_left"))
		return
	}

	for i := 0; i <= 3; i++ {
		templates := []string{}
		for key := range ru {
			if strings.HasPrefix(key, fmt.Sprintf("faggot_game_%d", i)) {
				templates = append(templates, key)
			}
		}
		template := templates[rand.Intn(len(templates))]
		phrase := i18n(template)
		logger.Infof("%d Using template: %s", m.Chat.ID, template)

		if i == 3 {
			phrase = fmt.Sprintf(phrase, winner.mention())
		}

		g.reply(m, phrase)

		r := rand.Intn(3) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	// Insert into DB after reply to prevent faggot_winner_known invocation by multiple /pidor calling
	entry = Entry{m.Chat.ID, winner.UserID, day, winner.Username}
	result := db.Create(&entry)
	logger.Infof("%d rows added: %d", m.Chat.ID, result.RowsAffected)
}

// Statistics for all time
func (g *Game) all(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_all_top"), ""}

	type result struct {
		Username string
		Count    int64
	}
	var results []result
	db.Table("faggot_entries").Select("username, count(*) as Count").Where("chat_id = ?", m.Chat.ID).Group("username").Order("Count desc").Scan(&results)

	if len(results) == 0 {
		logger.Warningf("%d No results for %s", m.Chat.ID, m.Text)
		return // Do not respond if there are no games yet
	}

	for i, res := range results {
		s = append(s, fmt.Sprintf(i18n("faggot_all_entry"), i+1, res.Username, res.Count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), len(results)))
	g.reply(m, strings.Join(s, "\n"))
}

// Current year statistics
func (g *Game) stats(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_stats_top"), ""}
	year := fmt.Sprintf("%d-%%", time.Now().Year())

	type result struct {
		Username string
		Count    int64
	}
	var results []result
	db.Table("faggot_entries").Select("username, count(*) as Count").Where("chat_id = ? AND day LIKE ?", m.Chat.ID, year).Group("username").Order("Count desc").Scan(&results)

	if len(results) == 0 {
		logger.Warningf("%d No results for %s", m.Chat.ID, m.Text)
		return // Do not respond if there are no games yet
	}

	for i, res := range results {
		s = append(s, fmt.Sprintf(i18n("faggot_stats_entry"), i+1, res.Username, res.Count))
	}

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), len(results)))
	g.reply(m, strings.Join(s, "\n"))
}

// Personal stat
func (g *Game) me(m *tb.Message) {
	if m.Private() {
		g.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	var count int64
	db.Model(&Entry{}).Where("chat_id = ? AND user_id = ?", m.Chat.ID, m.Sender.ID).Count(&count)

	me := Player{m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode}
	g.reply(m, fmt.Sprintf(i18n("faggot_me"), me.mention(), count))
}
