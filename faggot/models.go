package faggot

// Entry represents one day game result
type Entry struct {
	ChatID   int64
	UserID   int
	Day      string `gorm:"primaryKey"`
	Username string
}

// TableName ...
func (Entry) TableName() string {
	return "faggot_entries"
}

// Player represents game player
type Player struct {
	ChatID       int64 `gorm:"primaryKey"`
	UserID       int   `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

// TableName ...
func (Player) TableName() string {
	return "faggot_players"
}

func (u *Player) mention() string {
	if len(u.Username) > 0 {
		return "@" + u.Username
	}
	return u.FirstName + " " + u.LastName
}
