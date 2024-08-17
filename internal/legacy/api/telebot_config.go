package api

type TelebotConfigOption func(*TelebotConfig)

func WithBotToken(token string) TelebotConfigOption {
	return func(cfg *TelebotConfig) {
		cfg.BotToken = token
	}
}

func WithReportChatId(chatId int64) TelebotConfigOption {
	return func(cfg *TelebotConfig) {
		cfg.ReportChatId = chatId
	}
}

func WithBotAPIUrl(url string) TelebotConfigOption {
	return func(cfg *TelebotConfig) {
		cfg.BotAPIUrl = url
	}
}

type TelebotConfig struct {
	BotToken     string
	BotAPIUrl    string
	ReportChatId int64
}
