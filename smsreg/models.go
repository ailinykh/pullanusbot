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

func (s *Service) price() float32 {
	priceMap := map[string]float32{
		// "aol": 3,
		"gmail":      9,
		"facebook":   4,
		"mailru":     3,
		"vk":         45,
		"classmates": 4,
		"twitter":    3,
		// "mamba":      2,
		// "uber":       3,
		"telegram": 8,
		// "badoo":      5,
		// "drugvokrug": 5,
		"avito": 6,
		// "olx":        19,
		// "steam":      4,
		// "fotostrana": 4,
		"microsoft": 4,
		"viber":     8,
		"whatsapp":  9,
		// "wechat":    15,
		// "seosprint": 3,
		"instagram": 6,
		// "yahoo":     3,
		// "lineme":    3,
		// "kakaotalk": 5,
		// "meetme":    9,
		// "tinder":    3,
		// "nimses":    5,
		// "youla":     4,
		"other": 7,
	}
	if val, ok := priceMap[s.ID]; ok {
		return val
	}
	return 0
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
