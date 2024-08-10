package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func NewRebootServerFlow(serverApi core.ServerAPI, logger core.ILogger) *RebootServerFlow {
	return &RebootServerFlow{
		logger:       logger,
		serverApi:    serverApi,
		rebootLog:    []*logEntry{},
		cancelReboot: make(chan bool),
	}
}

type logEntry struct {
	timestamp time.Time
	message   string
}

const cancelRebootVpnButtonId = "cancel_reboot_vpn"

type RebootServerFlow struct {
	logger       core.ILogger
	serverApi    core.ServerAPI
	rebootLog    []*logEntry
	cancelReboot chan bool
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *RebootServerFlow) HandleText(message *core.Message, bot core.IBot) error {
	if message.Text != "/rebootserver" {
		return fmt.Errorf("not implemented")
	}

	return flow.Reboot(message, bot)
}

func (flow *RebootServerFlow) Reboot(message *core.Message, bot core.IBot) error {
	flow.logger.Infof("attempt to reboot server via %+v", message.Sender)

	for _, entry := range flow.rebootLog {
		if time.Since(entry.timestamp) < 5*time.Minute {
			flow.logger.Infof("reboot server timeout %s", entry.timestamp)
			text := fmt.Sprintf("Server already restarted at <b>%s</b>", entry.timestamp.Format(time.UnixDate))
			_, err := bot.SendText(text)
			return err
		}
	}

	entry := logEntry{
		timestamp: time.Now(),
		message:   message.Sender.FirstName + " " + message.Sender.LastName,
	}
	flow.rebootLog = append(flow.rebootLog, &entry)
	// Limit log size
	if len(flow.rebootLog) > 5 {
		flow.rebootLog = flow.rebootLog[1:]
	}

	ctx := context.Background()
	servers, err := flow.serverApi.GetServers(ctx)
	if err != nil {
		return err
	}

	if len(servers) != 1 {
		_, err = bot.SendText("Unexpected servers count. Sorry.")
		return err
	}

	button := &core.Button{Text: "❌ Cancel", ID: cancelRebootVpnButtonId}
	text := fmt.Sprintf("Restarting <b>%s</b> in <i>5 seconds...</i>\n\nTo cancel operation press the button below", servers[0].Name)
	sent, err := bot.SendText(text, [][]*core.Button{{button}})
	if err != nil {
		return err
	}

	select {
	case <-flow.cancelReboot:
		flow.rebootLog = flow.rebootLog[:len(flow.rebootLog)-1]
		return nil
	case <-time.After(5 * time.Second):
		flow.logger.Infof("restarting server due to cancel timeout reached")
		messages := []string{
			"✅ Server restarted.",
			"",
			"restart log:",
		}
		for _, entry := range flow.rebootLog {
			text := fmt.Sprintf("<i>%s by %s</i>", entry.timestamp.Format(time.UnixDate), entry.message)
			messages = append(messages, text)
		}
		_, err = bot.Edit(sent, strings.Join(messages, "\n"))
		if err != nil {
			return err
		}
		return flow.serverApi.RebootServer(ctx, servers[0])
	}
}

// GetButtonIds is a core.IButtonHandler protocol implementation
func (flow *RebootServerFlow) GetButtonIds() []string {
	return []string{cancelRebootVpnButtonId}
}

// ButtonPressed is a core.IButtonHandler protocol implementation
func (flow *RebootServerFlow) ButtonPressed(button *core.Button, message *core.Message, user *core.User, bot core.IBot) error {
	text := fmt.Sprintf("❌ Reboot interrupted by %s %s", user.FirstName, user.LastName)
	flow.logger.Info(text)
	_, err := bot.Edit(message, text)
	flow.cancelReboot <- true
	return err
}
