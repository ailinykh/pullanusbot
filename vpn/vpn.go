package vpn

import (
	"fmt"
	"os"
	"os/exec"
	i "pullanusbot/interfaces"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
	db  *gorm.DB

	btnCreate        = tb.InlineButton{Unique: "vpn_new_key", Text: i18n("vpn_new_key")}
	btnChooseMobile  = tb.InlineButton{Unique: "vpn_new_key_mobile", Text: i18n("vpn_new_key_choose_device_mobile"), Data: "Mobile"}
	btnChooseLaptop  = tb.InlineButton{Unique: "vpn_new_key_laptop", Text: i18n("vpn_new_key_choose_device_laptop"), Data: "Laptop"}
	btnChooseDesktop = tb.InlineButton{Unique: "vpn_new_key_desktop", Text: i18n("vpn_new_key_choose_device_desktop"), Data: "Desktop"}
	btnManage        = tb.InlineButton{Unique: "vpn_manage_keys", Text: i18n("vpn_manage_keys")}
	btnCancel        = tb.InlineButton{Unique: "vpn_cancel", Text: i18n("vpn_cancel"), Data: "cancel"}
)

// Vpn upload images to telegra.ph
type Vpn struct {
}

// Setup all nesessary command handlers
func (v *Vpn) Setup(b i.Bot, conn *gorm.DB) {
	bot, db = b, conn
	db.AutoMigrate(&VpnKey{})
	logger.Info("successfully initialized")

	bot.Handle("/vpnhelp", v.start)

	bot.Handle(&btnCreate, v.handleCreateKey)
	bot.Handle(&btnManage, v.handleManageKeys)
	bot.Handle(&btnCancel, v.handleCancel)

	bot.Handle(&btnChooseMobile, v.handleCreateKeyDeviceChoosen)
	bot.Handle(&btnChooseLaptop, v.handleCreateKeyDeviceChoosen)
	bot.Handle(&btnChooseDesktop, v.handleCreateKeyDeviceChoosen)
}

func (v *Vpn) start(m *tb.Message) {
	if !m.Private() {
		logger.Errorf("Not available for %#v", m.Chat)
		return
	}

	var count int64
	db.Model(&VpnKey{}).Where("user_id = ?", m.Sender.ID).Count(&count)

	var keyboard [][]tb.InlineButton

	switch true {
	case count == 0:
		keyboard = [][]tb.InlineButton{{btnCreate}, {btnCancel}}
	case count < 10:
		keyboard = [][]tb.InlineButton{{btnCreate}, {btnManage}, {btnCancel}}
	default:
		keyboard = [][]tb.InlineButton{{btnManage}, {btnCancel}}
	}

	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu, DisableWebPagePreview: true}
	bot.Send(m.Chat, i18n("vpn_start"), opts)
}

func (v *Vpn) handleCreateKey(c *tb.Callback) {
	keyboard := [][]tb.InlineButton{{btnChooseMobile}, {btnChooseLaptop}, {btnChooseDesktop}, {btnCancel}}
	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ReplyMarkup: menu}
	_, err := bot.Edit(c.Message, i18n("vpn_new_key_choose_device"), opts)
	if err != nil {
		logger.Error(err)
	}
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) handleCreateKeyDeviceChoosen(c *tb.Callback) {
	ts := time.Now()
	tsRFC := ts.Format(time.RFC3339)
	tsUnix := ts.Unix()

	cmd := fmt.Sprintf("easyrsa build-client-full %d-%d nopass", c.Sender.ID, tsUnix)
	out, err := executeVPNCommand(cmd)
	if err != nil {
		logger.Error(string(out))
		logger.Error(err)
		return
	}

	device := getAvailableDeviceName(c.Sender.ID, c.Data)
	logger.Infof("Creating new key %s for %d", device, c.Sender.ID)

	cmd = fmt.Sprintf("ovpn_getclient %d-%d", c.Sender.ID, tsUnix)
	out, err = executeVPNCommand(cmd)
	if err != nil {
		logger.Error(string(out))
		logger.Error(err)
		return
	}

	key := VpnKey{c.Sender.ID, tsRFC, device, c.Sender.Username, out}
	db.Create(&key)
	logger.Infof("New key %s for %d created", device, c.Sender.ID)

	v.sendKeyAndComplete(c.Message, key)
	bot.Respond(c, &tb.CallbackResponse{})

	adminChatID, err := strconv.ParseInt(os.Getenv("VPN_ADMIN_CHAT_ID"), 10, 64)
	if err != nil {
		logger.Error("VPN_ADMIN_CHAT_ID not set!")
		return
	}
	chat := &tb.Chat{ID: adminChatID}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown}
	bot.Send(chat, i18n("vpn_new_key_created_report", device, fmt.Sprintf("*%s %s* (%s)", c.Sender.FirstName, c.Sender.LastName, c.Sender.Username)), opts)
}

func (v *Vpn) handleManageKeys(c *tb.Callback) {
	var keyboard = [][]tb.InlineButton{}
	var keys []VpnKey
	db.Where("user_id = ?", c.Sender.ID).Find(&keys)

	for _, key := range keys {
		unique := fmt.Sprintf("manage_%d_%d", c.Sender.ID, key.unixTime())

		btn := tb.InlineButton{Unique: unique, Text: "🔑 " + key.Device + " (" + key.humanReadableTime() + ")", Data: unique}
		bot.Handle(&btn, v.handleManageKey)
		keyboard = append(keyboard, []tb.InlineButton{btn})
	}
	keyboard = append(keyboard, []tb.InlineButton{btnCancel})
	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
	_, err := bot.Edit(c.Message, i18n("vpn_manage_keys_choose"), opts)
	if err != nil {
		logger.Error(err)
	}
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) handleManageKey(c *tb.Callback) {
	key := parseKey(c.Data)

	uniqSend := fmt.Sprintf("send_%d_%d", c.Sender.ID, key.unixTime())
	btnSend := tb.InlineButton{Unique: uniqSend, Text: i18n("vpn_download_key"), Data: uniqSend}
	bot.Handle(&btnSend, v.handleSendKey)

	uniqRemove := fmt.Sprintf("remove_%d_%d", c.Sender.ID, key.unixTime())
	btnRemove := tb.InlineButton{Unique: uniqRemove, Text: i18n("vpn_remove_key"), Data: uniqRemove}
	bot.Handle(&btnRemove, v.handleRemoveKey)

	keyboard := [][]tb.InlineButton{{btnSend}, {btnRemove}, {btnCancel}}

	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
	_, err := bot.Edit(c.Message, i18n("vpn_manage_keys_choosen", key.Device, key.humanReadableTime()), opts)
	if err != nil {
		logger.Error(err)
	}
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) handleSendKey(c *tb.Callback) {
	key := parseKey(c.Data)

	v.sendKeyAndComplete(c.Message, key)
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) handleRemoveKey(c *tb.Callback) {
	key := parseKey(c.Data)

	btnRemoveConfirmation := tb.InlineButton{Unique: "confirm_" + c.Data, Text: i18n("vpn_remove_key"), Data: c.Data}
	bot.Handle(&btnRemoveConfirmation, v.handleRemoveKeyConfirmation)
	keyboard := [][]tb.InlineButton{{btnRemoveConfirmation}, {btnCancel}}
	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeMarkdown, ReplyMarkup: menu}
	_, err := bot.Edit(c.Message, i18n("vpn_remove_key_confirmation", key.Device, key.humanReadableTime()), opts)
	if err != nil {
		logger.Error(err)
	}
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) handleRemoveKeyConfirmation(c *tb.Callback) {
	key := parseKey(c.Data)
	logger.Infof("Removing key %s for %d", key.Device, key.UserID)

	cmd := fmt.Sprintf(`bash -c "echo 'yes' | ovpn_revokeclient %d-%d remove"`, c.Sender.ID, key.unixTime())
	out, err := executeVPNCommand(cmd)
	if err != nil {
		logger.Error(string(out))
		logger.Error(err)
		return
	}

	db.Delete(&key)
	logger.Infof("Key %s for %d removed", key.Device, key.UserID)

	_, err = bot.Edit(c.Message, i18n("vpn_remove_key_completed"))
	if err != nil {
		logger.Error(err)
	}
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) handleCancel(c *tb.Callback) {
	_, err := bot.Edit(c.Message, i18n("vpn_operation_canceled"))
	if err != nil {
		logger.Error(err)
	}
	bot.Respond(c, &tb.CallbackResponse{})
}

func (v *Vpn) sendKeyAndComplete(m *tb.Message, key VpnKey) {
	filename := fmt.Sprintf("%s-%s.ovpn", strings.ReplaceAll(m.Chat.Username, "_", "\\_"), key.Device)
	filepath := os.TempDir() + filename
	err := os.WriteFile(filepath, key.KeyData, 0644)
	defer os.Remove(filepath)
	if err != nil {
		logger.Error(err)
		return
	}

	doc := tb.Document{File: tb.FromDisk(filepath), FileName: filename}
	_, err = doc.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{})
	if err != nil {
		logger.Error(err)
	}

	_, err = bot.Edit(m, i18n("vpn_new_key_sent"))
	if err != nil {
		logger.Error(err)
	}
}

func getAvailableDeviceName(userID int, device string) string {
	var count int64
	var key VpnKey
	db.Last(&key).Where("user_id = ? AND device LIKE ?", userID, device+"%").Count(&count)

	if count == 0 {
		return device
	}

	parts := strings.Split(key.Device, "-")

	if len(parts) < 2 {
		return device + "-1"
	}

	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Error(err)
	}

	return fmt.Sprintf("%s-%d", device, idx+1)
}

func parseKey(data string) VpnKey {
	d := strings.Split(data, "_")
	userID, keyIdx := d[1], d[2]

	ts, err := strconv.ParseInt(keyIdx, 10, 64)
	if err != nil {
		logger.Error(err)
	}
	tm := time.Unix(ts, 0)

	var key VpnKey
	db.First(&key, "user_id = ? AND date = ?", userID, tm.Format(time.RFC3339))

	return key
}

func executeVPNCommand(cmd string) ([]byte, error) {
	debug, _ := strconv.ParseBool(os.Getenv("DEV"))
	var command string
	if debug {
		command = fmt.Sprintf(`docker-compose exec -T openvpn %s`, cmd)
	} else {
		command = fmt.Sprintf(`ssh -o StrictHostKeyChecking=no openvpn %s`, cmd)
	}
	return exec.Command("/bin/sh", "-c", command).CombinedOutput()
}
