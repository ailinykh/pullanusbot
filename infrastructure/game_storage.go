package infrastructure

import (
	"log"
	"path"

	"github.com/ailinykh/pullanusbot/v2/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func CreateGameStorage(gameID int64, factory IPlayerFactory) GameStorage {
	dbFile := path.Join(".", "pullanusbot.db")
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal(err)
	}
	conn.AutoMigrate(&Player{}, &Round{})

	s := GameStorage{conn, gameID, factory}
	return s
}

type GameStorage struct {
	conn          *gorm.DB
	gameID        int64
	playerFactory IPlayerFactory
}

func (db *GameStorage) GetPlayers() ([]*core.User, error) {
	var dbPlayers []Player
	var corePlayers []*core.User
	db.conn.Where("game_id = ?", db.gameID).Find(&dbPlayers)
	for _, p := range dbPlayers {
		corePlayers = append(corePlayers, &core.User{Username: p.Username})
	}
	return corePlayers, nil
}

func (s *GameStorage) GetRounds() ([]*core.Round, error) {
	var dbRounds []Round
	var coreRounds []*core.Round
	s.conn.Where("game_id = ?", s.gameID).Find(&dbRounds)
	for _, r := range dbRounds {
		player := &core.User{Username: r.Username}
		coreRounds = append(coreRounds, &core.Round{Day: r.Day, Winner: player})
	}
	return coreRounds, nil
}

func (s *GameStorage) AddPlayer(player *core.User) error {
	dbPlayer := s.playerFactory.CreatePlayer(player.Username)
	s.conn.Create(&dbPlayer)
	return nil
}

func (s *GameStorage) AddRound(round *core.Round) error {
	player := s.playerFactory.CreatePlayer(round.Winner.Username)
	dbRound := Round{
		GameID:   s.gameID,
		UserID:   player.UserID,
		Day:      round.Day,
		Username: round.Winner.Username,
	}
	s.conn.Create(&dbRound)
	return nil
}
