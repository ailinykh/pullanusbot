package smsreg

import (
	"fmt"
	"os"
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
		if service.enabled() {
			lastIdx := len(keyboard) - 1
			// nextBtn := tb.InlineButton{Unique: service.ID, Text: service.Title, Data: fmt.Sprintf("%s|%s|%s", balance.Amount, service.ID, service.Title)}
			nextBtn := tb.InlineButton{Unique: service.ID, Text: service.Title, Data: service.ID}
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
	bot.Send(m.Chat, fmt.Sprintf(i18n("sms_choose_service"), balance.Amount), opts)
}

func (s *SmsReg) handleService(c *tb.Callback) {
	logger.Infof("Confirm service %s", c.Data)
	btn := tb.InlineButton{Unique: "sms_get_number_btn", Text: i18n("sms_get_number_btn"), Data: c.Data}
	keyboard := [][]tb.InlineButton{[]tb.InlineButton{btn}}
	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
	bot.Edit(c.Message, fmt.Sprintf(i18n("sms_confirm_service"), c.Data), opts)
	bot.Handle(&btn, s.handleNumber)
	bot.Respond(c, &tb.CallbackResponse{})
}

func (s *SmsReg) handleNumber(c *tb.Callback) {
	logger.Infof("Requesting number for service: %s", c.Data)
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	bot.Edit(c.Message, fmt.Sprintf(i18n("sms_number_requested"), c.Data), opts)
	bot.Respond(c, &tb.CallbackResponse{})
	num, _ := client.getNum(c.Data)
	loop := 0

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
			btn := tb.InlineButton{Unique: "sms_number_ready_btn", Text: i18n("sms_number_ready_btn"), Data: num.ID}
			keyboard := [][]tb.InlineButton{[]tb.InlineButton{btn}}
			menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
			opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
			bot.Handle(&btn, s.handleSms)
			bot.Send(c.Message.Chat, fmt.Sprintf(i18n("sms_number_received"), tz.Number), opts)
			return
		}

		if tz.Base.Response != "TZ_INPOOL" {
			bot.Send(c.Message.Chat, "unknown response: "+tz.Base.Response)
			return
		}

		bot.Edit(c.Message, fmt.Sprintf(i18n("sms_number_requested_sec"), c.Data, loop*sleepTimeout), opts)
	}
}

func (s *SmsReg) handleSms(c *tb.Callback) {
	logger.Infof("Awaiting for sms for tzid: %s", c.Data)
	_, _ = client.setReady(c.Data)
	tz, _ := client.getState(c.Data)
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	bot.Edit(c.Message, fmt.Sprintf(i18n("sms_await_for_message"), tz.Number), opts)
	bot.Respond(c, &tb.CallbackResponse{})
	loop := 0
	shouldSetupKeyboard := true

	for {
		loop++
		time.Sleep(sleepTimeout * time.Second)
		tz, _ := client.getState(c.Data)
		logger.Info(tz)

		if shouldSetupKeyboard {
			shouldSetupKeyboard = false
			btnUsed := tb.InlineButton{Unique: "sms_feedback_used_btn", Text: i18n("sms_feedback_used_btn"), Data: c.Data}
			keyboard := [][]tb.InlineButton{[]tb.InlineButton{btnUsed}}
			menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true, InlineKeyboard: keyboard}
			opts = &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
			bot.Handle(&btnUsed, s.handleFeedbackUsed)
		}
		// drop the keyboard on success
		if tz.Base.Response == "TZ_NUM_ANSWER" {
			opts = &tb.SendOptions{ParseMode: tb.ModeMarkdown}
		}

		bot.Edit(c.Message, fmt.Sprintf(i18n("sms_await_for_message_sec"), tz.Number, loop*sleepTimeout), opts)
		// success
		if tz.Base.Response == "TZ_NUM_ANSWER" {
			btnOkay := tb.InlineButton{Unique: "sms_feedback_okay_btn", Text: i18n("sms_feedback_okay_btn"), Data: c.Data}
			btnUsed := tb.InlineButton{Unique: "sms_feedback_used_btn", Text: i18n("sms_feedback_used_btn"), Data: c.Data}
			keyboard := [][]tb.InlineButton{[]tb.InlineButton{btnOkay}, []tb.InlineButton{btnUsed}}
			menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true, InlineKeyboard: keyboard}
			opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
			bot.Handle(&btnOkay, s.handleFeedbackOkay)
			bot.Handle(&btnUsed, s.handleFeedbackUsed)
			bot.Send(c.Message.Chat, fmt.Sprintf(i18n("sms_message_received"), tz.Number, tz.Service, tz.Message), opts)
			return
		}
		// timeout
		if tz.Base.Response == "TZ_OVER_NR" {
			bot.Edit(c.Message, fmt.Sprintf(i18n("sms_message_timeout"), tz.Number), opts)
			return
		}

		if tz.Base.Response != "TZ_NUM_WAIT" {
			bot.Send(c.Message.Chat, "unexpected response: "+tz.Base.Response)
			return
		}
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
