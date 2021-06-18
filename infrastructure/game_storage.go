package infrastructure

import (
	"log"
	"os"
	"path"

	"github.com/ailinykh/pullanusbot/v2/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var conn *gorm.DB

// CreateGameStorage is a default GameStorage factory
func CreateGameStorage(gameID int64, factory IPlayerFactory) GameStorage {
	if conn == nil {
		dbFile := path.Join(getWorkingDir(), "pullanusbot.db")
		var err error
		conn, err = gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			log.Fatal(err)
		}

		if conn.Migrator().HasTable(&Player{}) && conn.Migrator().HasColumn(&Player{}, "chat_id") {
			log.Println("Extendend migration")
			conn.Migrator().RenameColumn(&Player{}, "chat_id", "game_id")
			conn.Migrator().RenameTable("faggot_entries", "faggot_rounds")
			conn.Migrator().RenameColumn(&Round{}, "chat_id", "game_id")
		} else {
			log.Println("Default migration")
			conn.AutoMigrate(&Player{}, &Round{})
		}
	}

	s := GameStorage{conn, gameID, factory}
	return s
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

//TODO: duplicated code
func getWorkingDir() string {
	workingDir := os.Getenv("WORKING_DIR")
	if len(workingDir) == 0 {
		return "pullanusbot-data"
	}
	return workingDir
}
