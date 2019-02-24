package main

import (
	"strings"
	"testing"
)

func TestProxyCommandRespondsWithProxyInfo(t *testing.T) {
	defer tearUp(t)()
	info := Info{}
	info.initialize()
	m := getGroupMessage()

	info.proxy(m)

	text := bot.(*FakeBot).replyText()
	if !strings.Contains(text, "secret") {
		t.Log(text)
		t.Error("/proxy command must respond with proxy information")
	}
}
