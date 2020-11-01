package smsreg

import (
	"fmt"
	"runtime"
)

var ru = map[string]string{
	"sms_choose_service":        "💰 Текущий баланс: *%s*\n\nПожалуйста, выерите сервис:",
	"sms_confirm_service":       "Выбран сервис: *%s*\n\nДля подтверждения нажмите кнопку ниже",
	"sms_get_number_btn":        "📲 Получить номер",
	"sms_number_requested":      "Выбран сервис: *%s*\n\n_Запрашиваю номер..._",
	"sms_number_requested_sec":  "Выбран сервис: *%s*\n\n_Запрашиваю номер (%d сек)..._",
	"sms_number_not_found":      "К сожалению, в данный момент нет свободных номеров, попробуйте позднее",
	"sms_number_received":       "☎️ Ваш номер `%s`\n\nНажмите кнопку, когда смс будет отправлено",
	"sms_number_ready_btn":      "✅ Готов",
	"sms_await_for_message":     "☎️ Ваш номер `%s`\n\n_Ожидаем смс..._",
	"sms_await_for_message_sec": "☎️ Ваш номер `%s`\n\n_Ожидаем смс (%d сек)..._",
	"sms_message_timeout":       "☎️ Ваш номер `%s`\n\n⚠️ К сожалению, никаких сообщений так и не поступило. Попробуйте позднее.",
	"sms_message_received":      "💬 Получено сообщение для сервиса *%s* на номер *%s*\n\n_%s_",
	"sms_feedback_okay_btn":     "✅ Всё получилось",
	"sms_feedback_used_btn":     "❌ Номер уже использован",
	"sms_finished_text":         "Спасибо!",
}

func i18n(key string) string {

	if val, ok := ru[key]; ok {
		return val
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}
