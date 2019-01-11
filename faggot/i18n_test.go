package faggot

import (
	"strings"
	"testing"
)

func TestI18nMissedKey(t *testing.T) {
	var missedKey = "some_key_that_not_exists"
	var text = i18n(missedKey)

	if !strings.Contains(text, "KEY_MISSED") {
		t.Log(text)
		t.Error("i18n() must inform about missed key")
	}
}
