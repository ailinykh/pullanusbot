package infrastructure

type Player struct {
	GameID       int64 `gorm:"primaryKey"`
	UserID       int   `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

func (Player) TableName() string {
	return "faggot_players"
}
