package infrastructure

import (
	"fmt"
	"runtime"
)

func CreateVpnLocalizer() *VpnLocalizer {
	return &VpnLocalizer{
		map[string]map[string]string{"ru": {
			"vpn_button_create_key":              "🔑 Создать новый ключ",
			"vpn_button_manage_key":              "🔐 Управление ключами",
			"vpn_button_remove_key":              "❌ Удалить ключ",
			"vpn_button_back":                    "⏪ Назад",
			"vpn_button_cancel":                  "❌ Отмена",
			"vpn_enter_create_key_name":          "Придумайте <b>имя</b> для ключа.\nЭто может быть любой набор слов, который поможет вам понять, для чего вы используете тот или иной ключ.\n\nНапример:<i>\n- Мой ключ\n- Ключ для друзей\n- Родители</i>\n\nнапишите имя в следующем сообщении",
			"vpn_enter_create_key_name_too_long": "Давайте придумаем что-то более лаконичное",
			"vpn_enter_delete_key_name_top":      "Введите имя ключа, который хотите <b>удалить</b>\n",
			"vpn_enter_delete_key_name_item":     "<i>%s</i>",
			"vpn_key_created":                    "✅ Вы успешно создали новый ключ\n\n<code>%s</code>\n\nтеперь скопируйте ключ в буффер обмена (простым нажатием на него) и вставьте его в приложение",
			"vpn_key_deleted":                    "✅ Ключ \"<i>%s</i>\" удалён!\n\n",
			"vpn_key_not_found":                  "❌ Ключ не найден\n\n",
			"vpn_key_list_top":                   "🔑 Активные ключи:\n",
			"vpn_key_list_item":                  "<b>%d.</b> %s\n<code>%s</code>\n",
			"vpn_key_list_bottom":                "\nВсего ключей: <b>%d</b>",
			"vpn_welcome":                        "🌏 <b>VPN всего за 3 простых шага</b>\n\n1️⃣ Установите клиент <a href='https://getoutline.org/'>outline</a> на ваше устройство:\n\n📱 <a href='https://itunes.apple.com/us/app/outline-app/id1356177741'>iOS / iPhone / iPad</a>\n📱 <a href='https://play.google.com/store/apps/details?id=org.outline.android.client'>Android</a>\n\n🖥 <a href='https://itunes.apple.com/us/app/outline-app/id1356178125'>macOS</a>\n🪟 <a href='https://raw.githubusercontent.com/Jigsaw-Code/outline-releases/master/client/stable/Outline-Client.exe'>Windows</a>\n🐧 <a href='https://raw.githubusercontent.com/Jigsaw-Code/outline-releases/master/client/stable/Outline-Client.AppImage'>Linux</a>\n\n2️⃣ Нажмите на кнопку <i>\"Создать новый ключ\"</i>\n\n3️⃣ Скопируйте полученный ключ в клиент",
		}},
	}
}

// VpnLocalizer for faggot game
type VpnLocalizer struct {
	langs map[string]map[string]string
}

// I18n is a core.ILocalizer implementation
func (l *VpnLocalizer) I18n(lang, key string, args ...interface{}) string {
	if val, ok := l.langs[lang][key]; ok {
		return fmt.Sprintf(val, args...)
	}

	if val, ok := l.langs["ru"][key]; ok {
		return fmt.Sprintf(val, args...)
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}

// AllKeys is a core.ILocalizer implementation
func (l *VpnLocalizer) AllKeys() []string {
	keys := make([]string, 0, len(l.langs["ru"]))
	for k := range l.langs["ru"] {
		keys = append(keys, k)
	}
	return keys
}
