package main

import (
	"log"
	"os"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Vpn structure
type Vpn struct {
}

// initialize database and all nesessary command handlers
func (v *Vpn) initialize() {
	if os.Getenv("VPN_HOST") == "" {
		log.Println("VPN: VPN_HOST not set! Skipping...")
		return
	}

	log.Println("VPN: database initialization")

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS vpn_users (user_id INTEGER, enabled INTEGER, phone_number TEXT, first_name TEXT, last_name TEXT, username TEXT)")
	checkErr(err)

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS vpn_keys (user_id INTEGER, key_title TEXT, key_data TEXT)")
	checkErr(err)

	log.Println("VPN: subscribing to bot events")

	bot.Handle("/vpnhelp", v.help)
	bot.Handle(tb.OnContact, v.contact)
	bot.Handle("/test", v.test)

	log.Println("VPN: successfully initialized")
}

func (v *Vpn) reply(m *tb.Message, text string) {
	bot.Send(m.Chat, text, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (v *Vpn) help(m *tb.Message) {
	if !m.Private() {
		v.reply(m, i18n("vpn_not_available_for_groups"))
		return
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM vpn_users WHERE user_id = ?", m.Sender.ID).Scan(&count)
	checkErr(err)

	if count == 0 {
		replyBtn := tb.ReplyButton{Text: "Авторизоваться по номеру телефона", Contact: true}
		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{replyBtn},
		}

		bot.Send(m.Sender, i18n("vpn_info"), &tb.ReplyMarkup{
			ReplyKeyboard: replyKeys,
		})
		return
	}

	var phone string
	err = db.QueryRow("SELECT phone_number FROM vpn_users WHERE user_id = ?", m.Sender.ID).Scan(&phone)
	checkErr(err)
	log.Println("VPN: phone: " + phone)
}

func (v *Vpn) contact(m *tb.Message) {
	if m.Contact.UserID != m.Sender.ID {
		v.reply(m, i18n("vpn_fraud"))
		return
	}

	stmt, err := db.Prepare("INSERT INTO vpn_users(user_id, enabled, phone_number, first_name, last_name, username) values(?,?,?,?,?,?)")
	checkErr(err)
	defer stmt.Close()

	_, err = stmt.Exec(m.Sender.ID, 0, m.Contact.PhoneNumber, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username)
	checkErr(err)
	log.Printf("VPN: New user added %d %s", m.Sender.ID, m.Sender.Username)

	bot.Send(m.Sender, i18n("vpn_registered"), &tb.ReplyMarkup{ReplyKeyboardRemove: true})
}

func (v *Vpn) test(m *tb.Message) {
	// Do smth
}
