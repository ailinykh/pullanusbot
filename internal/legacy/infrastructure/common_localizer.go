package infrastructure

import (
	"fmt"
	"runtime"
)

func CreateCommonLocalizer() *CommonLocalizer {
	return &CommonLocalizer{
		map[string]map[string]string{"ru": {
			"start_welcome": "Привет!",
			"help": `Вот что я могу:
			
- видео, загруженное как файл, я сконвертирую в mp4 и отправлю обратно (до 20MB)
- если прислать мне сылку на видео, я скачаю его и загружу в этот же чат как видео
- ссылки на видео в <i>tiktok</i>, <i>twitter</i> и <i>instagram reels</i> так же поддерживаются
- у меня можно получить доступ к /proxy для telegram (на случай, если его опять заблокируют)
- в групповых чатах ролики на youtube длиною до <i>10 минут</i> я так же скачиваю и присылаю как видео
- если дать мне права на удаление сообщений, я буду удалять исходное сообщение с ссылкой
- в личном чате я могу скачать и прислать частями по 50MB любой ролик на youtube, достаточно просто прислать мне ссылку
- если прислать мне картинку, я загружу её на telegra.ph и отправлю ссылку в ответ
- функционал постоянно добавляется`,
		}, "en": {
			"start_welcome": "Welcome!",
			"help": `Here is what i can do:

- If you upload a video file, I will convert it to mp4 and send it back to you (up to 20MB).
- If you send me a http link to a video, I will upload it to this chat as a video file.
- Links to videos on <i>TikTok</i>, <i>Twitter</i>, and <i>Instagram Reels</i> are also supported.
- In group chats, I can download and send YouTube videos up to <i>10 minutes</i> long as video files.
- If you assign me an <i>admin</i> role in a group chat with <b>delete messages</b> permission, I will delete the original message with the link.
- In a private chat, I can download and send any YouTube video in parts of 50MB each; just send me the link.
- If you send me an image, I will upload it to telegra.ph and send the link in response.
- Functionality is constantly being added.
			`,
		}},
	}
}

// CommonLocalizer for faggot game
type CommonLocalizer struct {
	langs map[string]map[string]string
}

// I18n is a core.ILocalizer implementation
func (l *CommonLocalizer) I18n(lang, key string, args ...interface{}) string {
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
func (l *CommonLocalizer) AllKeys() []string {
	keys := make([]string, 0, len(l.langs["ru"]))
	for k := range l.langs["ru"] {
		keys = append(keys, k)
	}
	return keys
}
