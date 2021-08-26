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

// Round that can be persistent on disk
type Round struct {
	GameID   int64
	UserID   int
	Day      string `gorm:"primaryKey"`
	Username string
}

// TableName gorm API
func (Round) TableName() string {
	return "faggot_rounds"
}
