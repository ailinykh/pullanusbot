package infrastructure

// Player that can be persistent on disk
type Player struct {
	GameID       int64 `gorm:"primaryKey"`
	UserID       int   `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

// TableName gorm API
func (Player) TableName() string {
	return "faggot_players"
}
