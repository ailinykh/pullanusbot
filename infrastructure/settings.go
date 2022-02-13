package infrastructure

import "time"

type Settings struct {
	ChatID    int64
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}
