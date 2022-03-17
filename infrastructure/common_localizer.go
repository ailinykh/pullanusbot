package infrastructure

import (
	"fmt"
	"runtime"
)

func CreateCommonLocalizer() *CommonLocalizer {
	return &CommonLocalizer{
		map[string]map[string]string{"ru": {
			"start_welcome": `Привет! Вот что я могу:
			
- видео, загруженное как файл, я сконвертирую в mp4 и отправлю обратно (до 20MB)
- если прислать мне сылку на видео, я скачаю его и загружу в этот же чат как видео
- ссылки на видео в <i>tiktok</i>, <i>twitter</i> и <i>instagram reels</i> так же поддерживаются
- у меня можно получить доступ к proxy для telegram (на случай, если его опять заблокируют)
- в групповых чатах ролики на youtube длиною до 10 минут я так же скачиваю и присылаю как видео
- если дать мне права на удаление сообщений, я буду удалять исходное сообщение с ссылкой
- в личном чате я могу скачать и прислать частями по 50MB любой ролик на youtube, достаточно просто прислать мне ссылку
- если прислать мне картинку, я загружу её на telegra.ph и отправлю ссылку в ответ
- функционал постоянно добавляется`,
		}},
	}
}

// CommonLocalizer for faggot game
type CommonLocalizer struct {
	langs map[string]map[string]string
}

// I18n is a core.ILocalizer implementation
func (l *CommonLocalizer) I18n(key string, args ...interface{}) string {

	if val, ok := l.langs["ru"][key]; ok {
		return fmt.Sprintf(val, args...)
	}

	_, file, line, _ := runtime.Caller(0)
	return fmt.Sprintf("%s:%d KEY_MISSED:\"%s\"", file, line, key)
}

// AllKeys is a core.ILocalizer implementation
func (l *CommonLocalizer) AllKeys() []string {
	keys := make([]string, 0, len(ru))
	for k := range l.langs["ru"] {
		keys = append(keys, k)
	}
	return keys
}
