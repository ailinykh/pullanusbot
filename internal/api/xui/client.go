package xui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/google/uuid"
)

const inboundId = 4

func NewClient(l core.Logger, baseUrl, cookie string) *Client {
	headers := map[string]string{
		"Cookie": fmt.Sprintf("x-ui=%s;", cookie),
	}
	client := &http.Client{Transport: NewAddHeaderTransport(nil, headers)}
	return &Client{l, strings.TrimSuffix(baseUrl, "/"), client}
}

type Client struct {
	l       core.Logger
	baseUrl string
	client  *http.Client
}

// GetKeys is a core.IVpnAPI interface implementation
func (c *Client) GetKeys(chatId int64) ([]*legacy.VpnKey, error) {
	urlString := fmt.Sprintf("%s/xui/API/inbounds/get/%d", c.baseUrl, inboundId)
	resp, err := c.client.Get(urlString)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	var inboundResp InboundResponse
	err = json.Unmarshal(data, &inboundResp)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	var inboundStreamSettings InboundStreamSettings
	err = json.Unmarshal([]byte(inboundResp.Obj.StreamSettings), &inboundStreamSettings)
	if err != nil {
		c.l.Error("failed to parse 'streamSettings'", "error", err)
		return nil, err
	}

	var inboundSettings InboundSettings
	err = json.Unmarshal([]byte(inboundResp.Obj.Settings), &inboundSettings)
	if err != nil {
		c.l.Error("failed to parse 'settings'", "error", err)
		return nil, err
	}

	baseUrl, err := url.Parse(c.baseUrl)
	if err != nil {
		c.l.Error("failed to parse baseUrl", "error", err)
		return nil, err
	}

	var keys []*legacy.VpnKey
	for _, client := range inboundSettings.Clients {
		parts := strings.SplitN(client.Email, "|", 3)
		if len(parts) == 3 && strconv.FormatInt(chatId, 10) == parts[0] {
			key := fmt.Sprintf(
				"%s://%s@%s:%d?type=%s&security=%s&pbk=%s&fp=%s&sni=%s&sid=&spx=%s#%s",
				inboundResp.Obj.Protocol,
				client.ID,
				baseUrl.Hostname(),
				inboundResp.Obj.Port,
				inboundStreamSettings.Network,
				inboundStreamSettings.Security,
				inboundStreamSettings.RealitySettings.Settings.PublikKey,
				inboundStreamSettings.RealitySettings.Settings.Fingerprint,
				inboundStreamSettings.RealitySettings.ServerNames[0],
				inboundStreamSettings.RealitySettings.Settings.SpiderX,
				// https://go.dev/play/p/pOfrn-Wsq5
				(&url.URL{Path: parts[2]}).String(),
			)

			keys = append(keys, &legacy.VpnKey{
				ID:     client.ID,
				ChatID: chatId,
				Title:  parts[2],
				Key:    key,
			})
		}
	}

	return keys, nil
}

// CreateKey is a core.IVpnAPI interface implementation
func (c *Client) CreateKey(keyName string, chatId int64, user *legacy.User) (*legacy.VpnKey, error) {
	settings := InboundSettings{
		Clients: []InboundClient{{
			ID:    uuid.NewString(),
			Flow:  "",
			Email: fmt.Sprintf("%d|%s|%s", chatId, user.DisplayName(), keyName),
			TgId:  user.Username,
		}},
	}

	settingsData, err := json.Marshal(settings)
	if err != nil {
		c.l.Error("failed to marshal settings", "error", err)
		return nil, err
	}

	createClientReq := CreateClientRequest{
		ID:       inboundId,
		Settings: string(settingsData),
	}

	reqData, err := json.Marshal(createClientReq)
	if err != nil {
		c.l.Error("failed to marshal create client request", "error", err)
		return nil, err
	}

	urlString := fmt.Sprintf("%s/xui/API/inbounds/addClient", c.baseUrl)
	resp, err := c.client.Post(urlString, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		c.l.Error("failed create client", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.l.Error("failed read body", "error", err)
		return nil, err
	}

	var clientResp CreateClientResponse
	err = json.Unmarshal(data, &clientResp)
	if err != nil {
		c.l.Error("failed to unmarshal create client response", "error", err)
		return nil, err
	}

	if !clientResp.Success {
		c.l.Error("failed to create client", "error", clientResp.Msg)
		return nil, errors.New(clientResp.Msg)
	}

	keys, err := c.GetKeys(chatId)
	if err != nil {
		c.l.Error("failed to get keys", "error", err)
		return nil, err
	}

	if len(keys) < 1 {
		return nil, fmt.Errorf("expected at least one key, got %d", len(keys))
	}

	c.l.Info("key created", "user", user.DisplayName(), "name", keyName, "link", keys[len(keys)-1].Key)

	return keys[len(keys)-1], nil
}

// DeleteKey is a core.IVpnAPI interface implementation
func (c *Client) DeleteKey(key *legacy.VpnKey) error {
	urlString := fmt.Sprintf("%s/xui/API/inbounds/%d/delClient/%s", c.baseUrl, inboundId, key.ID)
	// content-type doesn't make sense
	// https://github.com/alireza0/x-ui/blob/3a7c00fc5feafdc8cd15b023113d489794bdbdcd/web/controller/inbound.go#L168
	_, err := c.client.Post(urlString, "application/json", nil)
	if err != nil {
		c.l.Error("failed to delete key", "error", err)
		return err
	}
	return nil
}
