package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func NewRebootServerFlow(serverApi core.ServerAPI, commandService core.ICommandService, logger core.ILogger, opts *RebootServerOptions) *RebootServerFlow {
	return &RebootServerFlow{
		logger:         logger,
		serverApi:      serverApi,
		commandService: commandService,
		rebootLog:      []*logEntry{},
		confirmReboot:  make(chan bool),
		opts:           opts,
	}
}

type logEntry struct {
	timestamp time.Time
	message   string
}

const confirmRebootVpnButtonId = "confirm_reboot_vpn"

type RebootServerOptions struct {
	ChatId  int64
	Command string
}

type RebootServerFlow struct {
	logger         core.ILogger
	serverApi      core.ServerAPI
	commandService core.ICommandService
	rebootLog      []*logEntry
	confirmReboot  chan bool
	opts           *RebootServerOptions
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *RebootServerFlow) HandleText(message *core.Message, bot core.IBot) error {
	if message.Text != flow.opts.Command {
		return fmt.Errorf("not implemented")
	}

	if message.Chat.ID != flow.opts.ChatId {
		return fmt.Errorf("wrong chat id. expect: %d, got: %d", flow.opts.ChatId, message.Chat.ID)
	}

	if err := flow.commandService.EnableCommands(flow.opts.ChatId, []core.Command{{
		Text:        flow.opts.Command,
		Description: "Reboot VPN Server",
	}}, bot); err != nil {
		return err
	}

	return flow.Reboot(message, bot)
}

func (flow *RebootServerFlow) Reboot(message *core.Message, bot core.IBot) error {
	flow.logger.Infof("attempt to reboot server via %+v", message.Sender)

	for _, entry := range flow.rebootLog {
		if time.Since(entry.timestamp) < 5*time.Minute {
			flow.logger.Infof("reboot server timeout %s", entry.timestamp)
			text := fmt.Sprintf("🔴 Server already restarted at <b>%s</b>", entry.timestamp.Format(time.UnixDate))
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

	button := &core.Button{
		ID:      confirmRebootVpnButtonId,
		Text:    "🟢 Yes, reboot",
		Payload: servers[0].Name,
	}
	messages := []string{
		fmt.Sprintf("🟡 Server reboot requested by %s %s.", message.Sender.FirstName, message.Sender.LastName),
		"",
		fmt.Sprintf("To reboot <b>%s</b> press the button below:", servers[0].Name),
	}
	sent, err := bot.SendText(strings.Join(messages, "\n"), [][]*core.Button{{button}})
	if err != nil {
		return err
	}

	select {
	case <-flow.confirmReboot:
		if err = flow.serverApi.RebootServer(ctx, servers[0]); err != nil {
			return err
		}

		go func() {
			time.Sleep(5 * time.Minute)
			messages := []string{
				"🟢 Server restarted.",
				"",
				"restart log:",
			}
			for _, entry := range flow.rebootLog {
				text := fmt.Sprintf("<i>%s by %s</i>", entry.timestamp.Format(time.UnixDate), entry.message)
				messages = append(messages, text)
			}
			_, err = bot.SendText(strings.Join(messages, "\n"))
			if err != nil {
				flow.logger.Error(err)
			}
		}()
	case <-time.After(5 * time.Second):
		flow.logger.Infof("server reboot interrupted due to confirm timeout reached")
		flow.rebootLog = flow.rebootLog[:len(flow.rebootLog)-1]

		_, err = bot.Edit(sent, "🔴 Server reboot canceled")
		return err
	}
	return err
}

// GetButtonIds is a core.IButtonHandler protocol implementation
func (flow *RebootServerFlow) GetButtonIds() []string {
	return []string{confirmRebootVpnButtonId}
}

// ButtonPressed is a core.IButtonHandler protocol implementation
func (flow *RebootServerFlow) ButtonPressed(button *core.Button, message *core.Message, user *core.User, bot core.IBot) error {
	flow.logger.Infof("server reboot confirmed by %d %s", user.ID, user.DisplayName())
	text := fmt.Sprintf("🟡 Server reboot in progress by %s %s...", user.FirstName, user.LastName)
	if _, err := bot.Edit(message, text); err != nil {
		return err
	}
	flow.confirmReboot <- true
	return nil
}
