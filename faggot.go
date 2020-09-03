package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Faggot structure
type Faggot struct {
}

// FaggotPlayer struct for serialization
type FaggotPlayer struct {
	*tb.User
}

func (p *FaggotPlayer) mention() string {
	return strings.Join([]string{"@", p.Username}, "")
}

// FaggotEntry struct for game result serialization
type FaggotEntry struct {
	Day      string `json:"day"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// FaggotStat is a game statistics structure
type FaggotStat struct {
	stat []FaggotPlayerStat
}

// FaggotPlayerStat just a simple stat representation
type FaggotPlayerStat struct {
	Player string
	Count  int
}

// Increment stat for given player name. Creates if not exists
func (s *FaggotStat) Increment(player string) {
	if s.stat == nil {
		s.stat = []FaggotPlayerStat{}
	}
	found := false
	for i, stat := range s.stat {
		if stat.Player == player {
			found = true
			s.stat[i].Count++
		}
	}
	if !found {
		s.stat = append(s.stat, FaggotPlayerStat{Player: player, Count: 1})
	}
}

func (s FaggotStat) Len() int {
	return len(s.stat)
}

func (s FaggotStat) Less(i, j int) bool {
	return s.stat[i].Count < s.stat[j].Count
}

func (s FaggotStat) Swap(i, j int) {
	foo := s.stat[i]
	s.stat[i] = s.stat[j]
	s.stat[j] = foo
}

// initialize database and all nesessary command handlers
func (f *Faggot) initialize() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS faggot_players (chat_id INTEGER, user_id INTEGER, first_name TEXT, last_name TEXT, username TEXT, language_code TEXT)")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS faggot_entries (day TEXT, chat_id INTEGER, user_id INTEGER, username TEXT)")
	checkErr(err)

	bot.Handle("/pidorules", f.rules)
	bot.Handle("/pidoreg", f.reg)
	bot.Handle("/pidor", f.play)
	bot.Handle("/pidorall", f.all)
	bot.Handle("/pidorstats", f.stats)
	bot.Handle("/pidorme", f.me)

	logger.Info("game started")
}

func (f *Faggot) reply(m *tb.Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeHTML})
}

// Print game rules
func (f *Faggot) rules(m *tb.Message) {
	f.reply(m, i18n("faggot_rules"))
}

// Register new player
func (f *Faggot) reg(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	logger.Infof("%d Registering new player", m.Chat.ID)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_players WHERE chat_id = ? AND user_id = ?", m.Chat.ID, m.Sender.ID).Scan(&count)
	checkErr(err)

	logger.Infof("%d Players found: %d", m.Chat.ID, count)

	if count > 0 {
		logger.Infof("%d Player already in game! (%d)", m.Chat.ID, m.Sender.ID)
		f.reply(m, i18n("faggot_already_in_game"))
		return
	}

	stmt, err := db.Prepare("INSERT INTO faggot_players(chat_id, user_id, first_name, last_name, username, language_code) values(?,?,?,?,?,?)")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(m.Chat.ID, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username, m.Sender.LanguageCode)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	logger.Infof("%d LastInsertId: (%d)", m.Chat.ID, id)

	logger.Infof("%d Player added to game (%d)", m.Chat.ID, m.Sender.ID)
	f.reply(m, i18n("faggot_added_to_game"))
}

// Active games sync
var activeGames = ConcurrentSlice{}

// Play POTD game
func (f *Faggot) play(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	logger.Infof("%d POTD: Playing pidor of the day for %s!", m.Chat.ID, day)

	rows, err := db.Query("SELECT user_id, username FROM faggot_players WHERE chat_id = ?", m.Chat.ID)
	checkErr(err)

	var players = []FaggotPlayer{}
	var userID int
	var username string
	for rows.Next() {
		err = rows.Scan(&userID, &username)
		checkErr(err)

		user := tb.User{ID: userID, Username: username}
		players = append(players, FaggotPlayer{User: &user})
	}

	logger.Infof("%d POTD: Players found: %d", m.Chat.ID, len(players))

	switch len(players) {
	case 0:
		logger.Infof("%d POTD: No players!", m.Chat.ID)
		player := FaggotPlayer{User: m.Sender}
		f.reply(m, fmt.Sprintf(i18n("faggot_no_players"), player.mention()))
		return
	case 1:
		logger.Infof("%d POTD: Not enough players!", m.Chat.ID)
		f.reply(m, i18n("faggot_not_enough_players"))
		return
	default:
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM faggot_entries WHERE chat_id = ? AND day = ?", m.Chat.ID, day).Scan(&count)
	checkErr(err)

	if count > 0 {
		err = db.QueryRow("SELECT username FROM faggot_entries WHERE chat_id = ? AND day = ?", m.Chat.ID, day).Scan(&username)
		checkErr(err)

		logger.Infof("%d POTD: Already known!", m.Chat.ID)

		phrase := fmt.Sprintf(i18n("faggot_winner_known"), username)
		f.reply(m, phrase)
		return
	}

	if activeGames.Index(m.Chat.ID) > -1 {
		logger.Infof("%d PODT: Game in progress! Do nothing!", m.Chat.ID)
		return
	}

	activeGames.Add(m.Chat.ID)
	defer activeGames.Remove(m.Chat.ID)

	winner := players[rand.Intn(len(players))]
	logger.Infof("%d POTD: Pidor of the day is %s!", m.Chat.ID, winner.Username)

	member, err := bot.ChatMemberOf(m.Chat, winner.User)
	logger.Infof("%d POTD: %v %v", m.Chat.ID, member, err)

	if member == nil ||
		member.Role == tb.Left ||
		member.Role == tb.Kicked {
		logger.Errorf("%d POTD: %v", m.Chat.ID, err)
		f.reply(m, i18n("faggot_winner_left"))
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
		logger.Infof("%d POTD: Using template: %s", m.Chat.ID, template)

		if i == 3 {
			phrase = fmt.Sprintf(phrase, winner.mention())
		}

		f.reply(m, phrase)

		r := rand.Intn(3) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	// Insert into DB after reply to prevent faggot_winner_known invocation by multiple /pidor calling
	stmt, err := db.Prepare("INSERT INTO faggot_entries(day, chat_id, user_id, username) values(?,?,?,?)")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(day, m.Chat.ID, winner.ID, winner.Username)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	logger.Infof("%d POTD: LastInsertId %d!", m.Chat.ID, id)
}

// Statistics for all time
func (f *Faggot) all(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_all_top"), ""}
	stats := FaggotStat{}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_entries WHERE chat_id = ?", m.Chat.ID).Scan(&count)
	checkErr(err)

	if count == 0 {
		return
	}

	rows, err := db.Query("SELECT username FROM faggot_entries WHERE chat_id = ?", m.Chat.ID)
	checkErr(err)

	var entries = []FaggotEntry{}
	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		checkErr(err)

		entries = append(entries, FaggotEntry{Username: username})
	}

	logger.Infof("%d All:  Entries found: %d", m.Chat.ID, len(entries))

	for _, entry := range entries {
		stats.Increment(entry.Username)
	}

	sort.Sort(sort.Reverse(stats))
	for i, stat := range stats.stat {
		s = append(s, fmt.Sprintf(i18n("faggot_all_entry"), i+1, stat.Player, stat.Count))
	}

	err = db.QueryRow("SELECT COUNT(*) FROM faggot_players WHERE chat_id = ?", m.Chat.ID).Scan(&count)
	checkErr(err)

	s = append(s, "", fmt.Sprintf(i18n("faggot_all_bottom"), count))
	f.reply(m, strings.Join(s, "\n"))
}

// Current year statistics
func (f *Faggot) stats(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	s := []string{i18n("faggot_stats_top"), ""}
	stats := FaggotStat{}
	loc, _ := time.LoadLocation("Europe/Zurich")
	currentYear := time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, loc)
	nextYear := time.Date(time.Now().Year()+1, time.January, 1, 0, 0, 0, 0, loc)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_entries WHERE chat_id = ?", m.Chat.ID).Scan(&count)
	checkErr(err)

	if count == 0 {
		logger.Infof("%d Stat: no entries. Exiting...", m.Chat.ID)
		return
	}

	rows, err := db.Query("SELECT day, username FROM faggot_entries WHERE chat_id = ?", m.Chat.ID)
	checkErr(err)

	var entries = []FaggotEntry{}
	var day string
	var username string
	for rows.Next() {
		err = rows.Scan(&day, &username)
		checkErr(err)

		entries = append(entries, FaggotEntry{Day: day, Username: username})
	}

	logger.Infof("%d Stat: Entries found: %d", m.Chat.ID, len(entries))

	for _, entry := range entries {
		t, _ := time.Parse("2006-01-02", entry.Day)

		if t.After(currentYear) && t.Before(nextYear) {
			stats.Increment(entry.Username)
		}
	}

	sort.Sort(sort.Reverse(stats))
	for i, stat := range stats.stat {
		s = append(s, fmt.Sprintf(i18n("faggot_stats_entry"), i+1, stat.Player, stat.Count))
	}

	err = db.QueryRow("SELECT COUNT(*) FROM faggot_players WHERE chat_id = ?", m.Chat.ID).Scan(&count)
	checkErr(err)

	s = append(s, "", fmt.Sprintf(i18n("faggot_stats_bottom"), count))
	f.reply(m, strings.Join(s, "\n"))
}

// Personal stat
func (f *Faggot) me(m *tb.Message) {
	if m.Private() {
		f.reply(m, i18n("faggot_not_available_for_private"))
		return
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM faggot_entries WHERE chat_id = ? AND user_id = ?", m.Chat.ID, m.Sender.ID).Scan(&count)
	checkErr(err)

	player := FaggotPlayer{User: m.Sender}
	f.reply(m, fmt.Sprintf(i18n("faggot_me"), player.mention(), count))
}
