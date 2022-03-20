package infrastructure

import "time"

type Settings struct {
	ChatID    int64 `gorm:"primaryKey"`
	Data      []byte
	CreatedAt time.Time `gorm:"autoUpdateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}
