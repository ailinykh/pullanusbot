package smsreg

import (
	"fmt"
	"os"
	"strings"
	"time"

	i "pullanusbot/interfaces"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

const sleepTimeout = 5

var (
	bot    i.Bot
	db     *gorm.DB
	client *Client
)

// SmsReg ...
type SmsReg struct {
}

// Setup all nesessary command handlers
func (s *SmsReg) Setup(b i.Bot, conn *gorm.DB) {
	key := os.Getenv("SMS_API_KEY")
	if len(key) == 0 {
		logger.Error("SMS_API_KEY required")
		return
	}

	bot = b
	db = conn
	client = NewClient(key)

	bot.Handle("/sms", s.start)
	logger.Info("Successfully initialized")
}

func (s *SmsReg) start(m *tb.Message) {
	balance, _ := client.getBalance()
	list, _ := client.getList()
	keyboard := [][]tb.InlineButton{}

	for _, service := range list.Services {
		if service.price() > 0 {
			lastIdx := len(keyboard) - 1
			nextBtn := tb.InlineButton{Unique: service.ID, Text: fmt.Sprintf("%s ₽%.02f", service.Title, service.price()), Data: fmt.Sprintf("%s|%s|%s", balance.Amount, service.ID, service.Title)}
			bot.Handle(&nextBtn, s.handleService)
			if lastIdx >= 0 && len(keyboard[lastIdx]) == 1 {
				keyboard[lastIdx] = append(keyboard[lastIdx], nextBtn)
			} else {
				keyboard = append(keyboard, []tb.InlineButton{nextBtn})
			}
		}
	}
	menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true, InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
	bot.Send(m.Chat, i18n("sms_choose_service", balance.Amount), opts)
}

func (s *SmsReg) handleService(c *tb.Callback) {
	logger.Infof("Confirm service step: %s", c.Data)
	data := strings.Split(c.Data, "|")
	balance, _, serviceTitle := data[0], data[1], data[2]
	btn := tb.InlineButton{Unique: "sms_get_number_btn", Text: i18n("sms_get_number_btn"), Data: c.Data}
	keyboard := [][]tb.InlineButton{[]tb.InlineButton{btn}}
	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
	bot.Edit(c.Message, i18n("sms_confirm_service", balance, serviceTitle), opts)
	bot.Handle(&btn, s.handleNumber)
	bot.Respond(c, &tb.CallbackResponse{})
}

func (s *SmsReg) handleNumber(c *tb.Callback) {
	logger.Infof("Requesting number step: %s", c.Data)
	data := strings.Split(c.Data, "|")
	balance, serviceID, serviceTitle := data[0], data[1], data[2]
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	bot.Edit(c.Message, i18n("sms_number_requested", balance, serviceTitle), opts)
	bot.Respond(c, &tb.CallbackResponse{})
	num, _ := client.getNum(serviceID)
	loop := 0

	if num.Base.Response == "ERROR" {
		if num.Base.Error == "WARNING_LOW_BALANCE" {
			bot.Send(c.Message.Chat, i18n("sms_balance_insufficient"))
			return
		}
	}

	for {
		loop++
		time.Sleep(sleepTimeout * time.Second)
		tz, _ := client.getState(num.ID)
		logger.Info(tz)
		if tz.Base.Response == "WARNING_NO_NUMS" {
			bot.Send(c.Message.Chat, i18n("sms_number_not_found"))
			return
		}

		if tz.Base.Response == "TZ_NUM_PREPARE" {
			btnReady := tb.InlineButton{Unique: "sms_number_ready_btn", Text: i18n("sms_number_ready_btn"), Data: c.Data + "|" + num.ID}
			btnUsed := tb.InlineButton{Unique: "sms_feedback_used_btn", Text: i18n("sms_feedback_used_btn"), Data: num.ID}
			keyboard := [][]tb.InlineButton{[]tb.InlineButton{btnReady}, []tb.InlineButton{btnUsed}}
			menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true, InlineKeyboard: keyboard}
			opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
			bot.Handle(&btnReady, s.handleSms)
			bot.Handle(&btnUsed, s.handleFeedbackUsed)
			bot.Edit(c.Message, i18n("sms_number_received", balance, serviceTitle, tz.Number), opts)
			return
		}

		if tz.Base.Response != "TZ_INPOOL" {
			logger.Error(tz)
			bot.Send(c.Message.Chat, "unknown response: "+tz.Base.Response)
			return
		}

		bot.Edit(c.Message, i18n("sms_number_requested_sec", balance, serviceTitle, loop*sleepTimeout), opts)
	}
}

func (s *SmsReg) handleSms(c *tb.Callback) {
	logger.Infof("Awaiting for sms step: %s", c.Data)
	data := strings.Split(c.Data, "|")
	balance, _, serviceTitle, tzid := data[0], data[1], data[2], data[3]
	_, _ = client.setReady(tzid)
	tz, _ := client.getState(tzid)
	number := tz.Number // For some reason TZ_NUM_ANSWER erases the number
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	bot.Edit(c.Message, i18n("sms_await_for_message", balance, serviceTitle, number), opts)
	bot.Respond(c, &tb.CallbackResponse{})
	loop := 0

	for {
		loop++
		time.Sleep(sleepTimeout * time.Second)
		tz, _ := client.getState(tzid)
		logger.Info(tz)
		// success
		if tz.Base.Response == "TZ_NUM_ANSWER" {
			btnOkay := tb.InlineButton{Unique: "sms_feedback_okay_btn", Text: i18n("sms_feedback_okay_btn"), Data: tzid}
			btnUsed := tb.InlineButton{Unique: "sms_feedback_used_btn", Text: i18n("sms_feedback_used_btn"), Data: tzid}
			keyboard := [][]tb.InlineButton{[]tb.InlineButton{btnOkay}, []tb.InlineButton{btnUsed}}
			menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true, InlineKeyboard: keyboard}
			opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
			bot.Handle(&btnOkay, s.handleFeedbackOkay)
			bot.Handle(&btnUsed, s.handleFeedbackUsed)
			bot.Edit(c.Message, i18n("sms_message_received", balance, serviceTitle, number, tz.Message), opts)
			return
		}
		// timeout
		if tz.Base.Response == "TZ_OVER_NR" {
			bot.Edit(c.Message, i18n("sms_message_timeout", balance, serviceTitle, number), opts)
			return
		}
		// Unexpected response
		if tz.Base.Response != "TZ_NUM_WAIT" {
			bot.Send(c.Message.Chat, "unexpected response: "+tz.Base.Response)
			return
		}
		// Status update
		bot.Edit(c.Message, i18n("sms_await_for_message_sec", balance, serviceTitle, number, loop*sleepTimeout), opts)
	}
}

func (s *SmsReg) handleFeedbackOkay(c *tb.Callback) {
	logger.Infof("Handle feedback Okay for tzid: %s", c.Data)
	bot.Edit(c.Message, &tb.ReplyMarkup{}) // remove keyboard
	bot.Respond(c, &tb.CallbackResponse{Text: i18n("sms_finished_text")})
	client.setOperationOkay(c.Data)
}

func (s *SmsReg) handleFeedbackUsed(c *tb.Callback) {
	logger.Infof("Handle feedback Used for tzid: %s", c.Data)
	bot.Edit(c.Message, &tb.ReplyMarkup{}) // remove keyboard
	bot.Respond(c, &tb.CallbackResponse{Text: i18n("sms_finished_text")})
	client.setOperationUsed(c.Data)
}
