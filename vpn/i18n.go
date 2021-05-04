package vpn

import (
	"fmt"
	"runtime"
)

var ru = map[string]string{
	"vpn_start":                         "1️⃣ *Установите OpenVPN*\n\n[Windows](https://openvpn.net/downloads/openvpn-connect-v3-windows.msi) / [Android](https://play.google.com/store/apps/details?id=net.openvpn.openvpn) / [iOS](https://itunes.apple.com/us/app/openvpn-connect/id590379981?mt=8) / [macOS](https://openvpn.net/downloads/openvpn-connect-v3-macos.dmg) / [Linux](https://openvpn.net/openvpn-client-for-linux/)\n\n2️⃣ *Создайте ключ и откройте его в приложении*",
	"vpn_new_key":                       "🔑 Создать новый ключ",
	"vpn_new_key_choose_device":         "🔑 Создать новый ключ\n\nВыберите устройство, для которого планируется ключ\n\nЭтот выбор ни на что не влияет, просто чтобы вам в будущем было легче различать ключи",
	"vpn_new_key_choose_device_mobile":  "📱 Мобильный телефон/планшет",
	"vpn_new_key_choose_device_laptop":  "💻 Ноутбук",
	"vpn_new_key_choose_device_desktop": "🖥 Персональный компьютер",
	"vpn_new_key_sent":                  "✅ Откройте ключ с помощью приложения OpenVPN",
	"vpn_new_key_created_report":        "💬 Новый VPN ключ _%s_ сгенерирован пользователем %s",
	"vpn_manage_keys":                   "⚙️ Управление ключами",
	"vpn_manage_keys_choose":            "Выберите ключ",
	"vpn_manage_keys_choosen":           "🔑 Выбран ключ *%s*, созданный _%s_",
	"vpn_download_key":                  "⬇️ Получить ключ",
	"vpn_remove_key":                    "❌ Удалить ключ",
	"vpn_remove_key_confirmation":       "‼️ Вы уверены, что хотите безвозвратно удалить ключ *%s* созданный _%s_?",
	"vpn_remove_key_completed":          "✅ Ключ удалён. Для продолжения нажмите /vpnhelp",
	"vpn_cancel":                        "Отмена",
	"vpn_operation_canceled":            "❌ Операция отменена. Для продолжения нажмите /vpnhelp",
}

func i18n(key string, args ...interface{}) string {

	if val, ok := ru[key]; ok {
		return fmt.Sprintf(val, args...)
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}
