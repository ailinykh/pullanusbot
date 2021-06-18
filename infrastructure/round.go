package infrastructure

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
