package smsreg

import (
	"fmt"
	"runtime"
)

var ru = map[string]string{
	"sms_balance_insufficient":  "❌ Необходимо пополнить баланс!",
	"sms_choose_service":        "💰 Баланс: *₽%s*\n\nПожалуйста, выерите сервис:",
	"sms_confirm_service":       "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n\nДля подтверждения нажмите кнопку ниже",
	"sms_get_number_btn":        "📲 Получить номер",
	"sms_number_requested":      "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n\n_Запрашиваю номер..._",
	"sms_number_requested_sec":  "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n\n_Запрашиваю номер (%d сек)..._",
	"sms_number_not_found":      "⚠️ К сожалению, в данный момент нет свободных номеров, попробуйте повторить операцию позже",
	"sms_number_received":       "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n☎️ Ваш номер `%s`\n\nНажмите кнопку, когда смс будет отправлено",
	"sms_number_ready_btn":      "✅ Готов",
	"sms_await_for_message":     "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n☎️ Ваш номер `%s`\n\n_Ожидаем смс..._",
	"sms_await_for_message_sec": "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n☎️ Ваш номер `%s`\n\n_Ожидаем смс (%d сек)..._",
	"sms_message_timeout":       "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n☎️ Ваш номер `%s`\n\n⚠️ К сожалению, сообщений так и не поступило",
	"sms_message_received":      "💰 Баланс: *₽%s*\n🔥 Сервис: *%s*\n☎️ Ваш номер `%s`\n💬 Сообщение:\n\n_%s_",
	"sms_feedback_okay_btn":     "✅ Всё получилось",
	"sms_feedback_used_btn":     "❌ Номер уже использован",
	"sms_finished_text":         "Спасибо!",
}

func i18n(key string, args ...interface{}) string {

	if val, ok := ru[key]; ok {
		return fmt.Sprintf(val, args...)
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}
