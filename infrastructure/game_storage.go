package infrastructure

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	"gorm.io/gorm"
)

// CreateGameStorage is a default GameStorage factory
func CreateGameStorage(conn *gorm.DB, gameID int64, factory IPlayerFactory, logger core.ILogger) *GameStorage {
	if conn.Migrator().HasTable(&Player{}) && conn.Migrator().HasColumn(&Player{}, "chat_id") {
		logger.Info("Extendend migration")
		conn.Migrator().RenameColumn(&Player{}, "chat_id", "game_id")
		conn.Migrator().RenameTable("faggot_entries", "faggot_rounds")
		conn.Migrator().RenameColumn(&Round{}, "chat_id", "game_id")
	} else {
		logger.Info("Default migration")
		conn.AutoMigrate(&Player{}, &Round{})
	}
	return &GameStorage{conn, gameID, factory}
}

// GameStorage implements core.IGameStorage interface
type GameStorage struct {
	conn          *gorm.DB
	gameID        int64
	playerFactory IPlayerFactory
}

// GetPlayers is a core.IGameStorage interface implementation
func (s *GameStorage) GetPlayers() ([]*core.User, error) {
	var dbPlayers []Player
	var corePlayers []*core.User
	s.conn.Where("game_id = ?", s.gameID).Find(&dbPlayers)
	for _, p := range dbPlayers {
		corePlayers = append(corePlayers, &core.User{Username: p.Username})
	}
	return corePlayers, nil
}

// GetRounds is a core.IGameStorage interface implementation
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

// AddPlayer is a core.IGameStorage interface implementation
func (s *GameStorage) AddPlayer(player *core.User) error {
	dbPlayer := s.playerFactory.CreatePlayer(player.Username)
	s.conn.Create(&dbPlayer)
	return nil
}

// AddRound is a core.IGameStorage interface implementation
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
