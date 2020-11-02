package smsreg

// Base is a basic response container
type Base struct {
	Response string `json:"response"`
	Error    string `json:"error_msg,omitempty"`
}

// Service represents social network / another service
// which you expect to receive sms from
type Service struct {
	ID    string `json:"service"`
	Title string `json:"description"`
}

func (s *Service) enabled() bool {
	enabled := []string{
		// "aol",
		"gmail",
		"facebook",
		"mailru",
		"vk",
		"classmates",
		"twitter",
		// "mamba",
		// "uber",
		"telegram",
		// "badoo",
		// "drugvokrug",
		"avito",
		// "olx",
		// "steam",
		// "fotostrana",
		"microsoft",
		"viber",
		"whatsapp",
		// "wechat",
		// "seosprint",
		"instagram",
		// "yahoo",
		// "lineme",
		// "kakaotalk",
		// "meetme",
		// "tinder",
		// "nimses",
		// "youla",
		"other",
	}
	for _, e := range enabled {
		if e == s.ID {
			return true
		}
	}
	return false
}

// Balance of current account
type Balance struct {
	Base
	Amount string `json:"balance"`
	User   string `json:"user"`
}

// List of available services
type List struct {
	Services []Service `json:"services"`
}

// Num ...
type Num struct {
	Base
	ID string `json:"tzid"`
}

// Tz ...
type Tz struct {
	Base
	Service string `json:"service"`
	Number  string `json:"number"`
	Message string `json:"msg"`
}
