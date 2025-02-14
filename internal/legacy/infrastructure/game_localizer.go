package infrastructure

import (
	"fmt"
	"runtime"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateGameLocalizer() core.ILocalizer {
	return &GameLocalizer{
		map[string]map[string]string{"ru": {
			// Faggot game
			"faggot_rules": `Правила игры <b>Пидор Дня</b> (только для групповых чатов):
	<b>1</b>. Зарегистрируйтесь в игру по команде /pidoreg
	<b>2</b>. Подождите пока зарегиструются все (или большинство :)
	<b>3</b>. Запустите розыгрыш по команде /pidor
	<b>4</b>. Просмотр статистики канала по команде /pidorstats, /pidorall
	<b>5</b>. Личная статистика по команде /pidorme
	<b>6</b>. Статистика за 2024 год по комнаде /pidor2024 (так же есть за 2023-2020)
	
	<b>Важно</b>, розыгрыш проходит только <b>раз в день</b>, повторная команда выведет <b>результат</b> игры.
	
	Сброс розыгрыша происходит каждый день в 12 часов ночи по UTC+2 (или два часа ночи по Москве).`,

			"faggot_not_available_for_private": "Извините, данная команда недоступна в личных чатах.",
			"faggot_added_to_game":             "<b>OK!</b> Ты теперь участвуешь в игре <b>\"Пидор Дня\"</b>",
			"faggot_info_updated":              "Профиль игрока обновлён!",
			"faggot_already_in_game":           "Эй! Ты уже в игре!",
			"faggot_no_players":                "Зарегистрированных в игру еще нет, а значит <b>пидор</b> ты - %s",
			"faggot_not_enough_players":        "Нужно как минимум два игрока, чтобы начать игру! Зарегистрируйся используя /pidoreg",
			"faggot_winner_known":              "Согласно моей информации, по результатам сегодняшнего розыгрыша <b>пидор дня</b> - %s!",
			"faggot_winner_left":               "Я нашел пидора дня, но похоже, что он вышел из этого чата (вот пидор!), так что попробуйте еще раз!",
			// 0
			"faggot_game_0_0": "Осторожно! <b>Пидор дня</b> активирован!",
			"faggot_game_0_1": "Система взломана. Нанесён урон. Запущено планирование контрмер.",
			"faggot_game_0_2": "Сейчас поколдуем...",
			"faggot_game_0_3": "Инициирую поиск <b>пидора дня</b>...",
			"faggot_game_0_4": "Итак... кто же сегодня <b>пидор дня</b>?",
			"faggot_game_0_5": "Кто сегодня счастливчик?",
			"faggot_game_0_6": "Зачем вы меня разбудили...",
			"faggot_game_0_7": "### RUNNING 'TYPIDOR.SH'...",
			"faggot_game_0_8": "Woop-woop! That's the sound of da pidor-police!",
			"faggot_game_0_9": "Опять в эти ваши игрульки играете? Ну ладно...",
			// 1
			"faggot_game_1_0": "<i>Шаманим-шаманим</i>...",
			"faggot_game_1_1": "<i>Где-же он</i>...",
			"faggot_game_1_2": "<i>Сканирую</i>...",
			"faggot_game_1_3": "<i>Военный спутник запущен, коды доступа внутри</i>...",
			"faggot_game_1_4": "<i>Хм</i>...",
			"faggot_game_1_5": "<i>Интересно</i>...",
			"faggot_game_1_6": "<i>Ведётся поиск в базе данных</i>...",
			"faggot_game_1_7": "<i>Машины выехали</i>",
			"faggot_game_1_8": "<i>(Ворчит) А могли бы на работе делом заниматься</i>",
			"faggot_game_1_9": "<i>Выезжаю на место...</i>",
			// 2
			"faggot_game_2_0": "Так-так, что же тут у нас...",
			"faggot_game_2_1": "КЕК!",
			"faggot_game_2_2": "Доступ получен. Аннулирование протокола.",
			"faggot_game_2_3": "Проверяю данные...",
			"faggot_game_2_4": "Ох...",
			"faggot_game_2_5": "Высокий приоритет мобильному юниту.",
			"faggot_game_2_6": "Ведётся захват подозреваемого...",
			"faggot_game_2_7": "Что с нами стало...",
			"faggot_game_2_8": "Сонно смотрит на бумаги",
			"faggot_game_2_9": "В этом совершенно нет смысла...",
			// 3
			"faggot_game_3_0": "Ого, вы посмотрите только! А <b>пидор дня</b> то - %s",
			"faggot_game_3_1": "Кажется, <b>пидор дня</b> - %s",
			"faggot_game_3_2": ` ​ .∧＿∧
	( ･ω･｡)つ━☆・*。
	⊂  ノ    ・゜+.
	しーＪ   °。+ *´¨)
			 .· ´¸.·*´¨)
			  (¸.·´ (¸.·"* ☆ ВЖУХ И ТЫ ПИДОР, %s`,
			"faggot_game_3_3": "И прекрасный человек дня сегодня... а нет, ошибка, всего-лишь <b>пидор</b> - %s",
			"faggot_game_3_4": "Анализ завершен. Ты <b>пидор</b>, %s",
			"faggot_game_3_5": "Ага! Поздравляю! Сегодня ты <b>пидор</b>, %s",
			"faggot_game_3_6": "Что? Где? Когда? А ты <b>пидор дня</b> - %s",
			"faggot_game_3_7": "Ну ты и <b>пидор</b>, %s",
			"faggot_game_3_8": "Кто бы мог подумать, но <b>пидор дня</b> - %s",
			"faggot_game_3_9": "Стоять! Не двигаться! Вы объявлены <b>пидором дня</b>, %s",

			"faggot_stats_top":      "Топ-10 <b>пидоров</b> за текущий год:",
			"faggot_stats_top_year": "Топ-10 <b>пидоров</b> за <b>%s</b> год:",
			"faggot_stats_entry":    "<b>%d</b>. %s — <i>%d раз(а)</i>",
			"faggot_stats_bottom":   "Всего участников — <i>%d</i>",
			"faggot_stats_empty":    "За <b>%s</b> год результатов не найдено",

			"faggot_all_top":    "Топ-10 <b>пидоров</b> за всё время:",
			"faggot_all_entry":  "<b>%d</b>. %s — <i>%d раз(а)</i>",
			"faggot_all_bottom": "Всего участников — <i>%d</i>",

			"faggot_me": "%s, ты был(а) <b>пидором дня</b> — %d раз!",
		}},
	}
}

// GameLocalizer for faggot game
type GameLocalizer struct {
	langs map[string]map[string]string
}

// I18n is a core.ILocalizer implementation
func (l *GameLocalizer) I18n(lang, key string, args ...interface{}) string {
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
func (l *GameLocalizer) AllKeys() []string {
	keys := make([]string, 0, len(l.langs["ru"]))
	for k := range l.langs["ru"] {
		keys = append(keys, k)
	}
	return keys
}
