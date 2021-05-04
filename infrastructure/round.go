package infrastructure

type Round struct {
	GameID   int64
	UserID   int
	Day      string `gorm:"primaryKey"`
	Username string
}

func (Round) TableName() string {
	return "faggot_rounds"
}
