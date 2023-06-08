package infrastructure

import (
	"fmt"
	"runtime"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateSettingsLocalizer() core.ILocalizer {
	return &SettingsLocalizer{
		map[string]map[string]string{"ru": {
			// basic
			"settings_canceled":      "✅ Operation completed",
			"settings_title":         "Available settings:",
			"settings_current_state": "Current state:",
			"settings_enabled":       "✅ enabled",
			"settings_disabled":      "❌ disabled",
			// settings
			"settings_instagram":          "📷 Instagram",
			"settings_link":               "🔗 Web links",
			"settings_tiktok":             "🎵 TikTok",
			"settings_twitter":            "🐦 Twitter",
			"settings_youtube":            "🎞 Youtube",
			"settings_flow":               "Process links",
			"settings_flow_remove_source": "Remove source message",
			// title
			"settings_title_instagram": "Instagram",
			"settings_title_link":      "Web links",
			"settings_title_tiktok":    "TikTok",
			"settings_title_twitter":   "Twitter",
			"settings_title_youtube":   "Youtube",
			// description
			"settings_description_instagram": "converts some instagram links into fullsize media with description",
			"settings_description_link":      "Web links",
			"settings_description_tiktok":    "TikTok",
			"settings_description_twitter":   "Twitter",
			"settings_description_youtube":   "Youtube",
			// buttons
			"settings_button_back":    "Back",
			"settings_button_cancel":  "❌ Cancel",
			"settings_button_enable":  "✅ Enable",
			"settings_button_disable": "❌ Disable",
		}},
	}
}

// SettingsLocalizer for settings flow
type SettingsLocalizer struct {
	langs map[string]map[string]string
}

// I18n is a core.ILocalizer implementation
func (l *SettingsLocalizer) I18n(key string, args ...interface{}) string {

	if val, ok := l.langs["ru"][key]; ok {
		return fmt.Sprintf(val, args...)
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}

// AllKeys is a core.ILocalizer implementation
func (l *SettingsLocalizer) AllKeys() []string {
	keys := make([]string, 0, len(ru))
	for k := range l.langs["ru"] {
		keys = append(keys, k)
	}
	return keys
}
