package faggot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"pullanusbot/utils"
	"sync"
	"testing"
	"time"

	"github.com/google/logger"
	"github.com/stretchr/testify/assert"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	messages []*tb.Message
	game     *Game
	user1    *tb.User
	user2    *tb.User
	chat     *tb.Chat
)

type Bot struct{}

func (Bot) ChatMemberOf(c *tb.Chat, u *tb.User) (*tb.ChatMember, error) {
	if c.ID == 43 {
		return nil, errors.New("api error: Bad Request: user not found")
	}
	return &tb.ChatMember{}, nil
}
func (Bot) Delete(tb.Editable) error                                               { return nil }
func (Bot) Download(*tb.File, string) error                                        { return nil }
func (Bot) Handle(interface{}, interface{})                                        {}
func (Bot) Notify(tb.Recipient, tb.ChatAction) error                               { return nil }
func (Bot) SendAlbum(tb.Recipient, tb.Album, ...interface{}) ([]tb.Message, error) { return nil, nil }
func (Bot) Start()                                                                 {}

func (Bot) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	m := &tb.Message{Text: what.(string)}
	messages = append(messages, m)
	return m, nil
}

func tearUp(t *testing.T) func() {
	// setup Logger
	l := logger.Init("pullanusbot_test", true, false, ioutil.Discard)
	// setup DB
	dbFile := path.Join(os.TempDir(), "pullanusbot-"+utils.RandStringRunes(4)+".db")
	logger.Info("Creating database: " + dbFile)
	db, _ = gorm.Open(sqlite.Open(dbFile+"?cache=shared"), nil)
	// setup Game
	messages = []*tb.Message{}
	user1 = &tb.User{ID: 1, FirstName: "Paul", LastName: "Durov", Username: "durov"}
	user2 = &tb.User{ID: 2, FirstName: "Nick", LastName: "Durov", Username: "durov2"}
	chat = &tb.Chat{ID: 42}
	game = &Game{}
	game.Setup(&Bot{}, db)
	return func() {
		l.Close()
		os.Remove(dbFile)
	}
}

func TestRulesCommand(t *testing.T) {
	defer tearUp(t)()
	game.rules(&tb.Message{Text: "/rules"})
	assert.Contains(t, messages[0].Text, "Правила игры")
}

func TestRegCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	chat := &tb.Chat{Type: tb.ChatPrivate}
	game.reg(&tb.Message{Text: "/reg", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "недоступна в личных")
}

func TestRegCommandAddsPlayerInGame(t *testing.T) {
	defer tearUp(t)()
	game.reg(&tb.Message{Text: "/reg", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "Ты в игре!")
}

func TestRegCommandAddsEachPlayerOnlyOnce(t *testing.T) {
	defer tearUp(t)()
	db.Create(&Player{chat.ID, user1.ID, user1.FirstName, user1.LastName, user1.Username, user1.LanguageCode})
	game.reg(&tb.Message{Text: "/reg", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "уже в игре!")
}

func TestPlayCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	chat := &tb.Chat{Type: tb.ChatPrivate}
	game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "недоступна в личных")
}

func TestPlayCommandRespondsNoPlayers(t *testing.T) {
	defer tearUp(t)()
	game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "Зарегистрированных в игру еще нет")
}

func TestPlayCommandRespondsNotEnoughPlayers(t *testing.T) {
	defer tearUp(t)()
	db.Create(&Player{chat.ID, user1.ID, user1.FirstName, user1.LastName, user1.Username, user1.LanguageCode})
	game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "Нужно как минимум два игрока")
}

func TestPlayCommandNotRespondsIfGameInProgress(t *testing.T) {
	defer tearUp(t)()
	db.Create(&Player{chat.ID, user1.ID, user1.FirstName, user1.LastName, user1.Username, user1.LanguageCode})
	db.Create(&Player{chat.ID, user2.ID, user2.FirstName, user2.LastName, user2.Username, user2.LanguageCode})

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
			wg.Done()
		}()
	}

	wg.Wait()
	assert.Equal(t, len(messages), 4)
}

func TestPlayCommandRespondsWinnerAlreadyKnown(t *testing.T) {
	defer tearUp(t)()
	day := time.Now().Format("2006-01-02")
	db.Create(&Player{chat.ID, user1.ID, user1.FirstName, user1.LastName, user1.Username, user1.LanguageCode})
	db.Create(&Player{chat.ID, user2.ID, user2.FirstName, user2.LastName, user2.Username, user2.LanguageCode})
	db.Create(&Entry{chat.ID, user1.ID, day, user1.Username})

	game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "по результатам сегодняшнего розыгрыша")
}

func TestPlayCommandRespondsUserIsNotMemberOfChat(t *testing.T) {
	defer tearUp(t)()
	chat := &tb.Chat{ID: 43}
	db.Create(&Player{chat.ID, user1.ID, user1.FirstName, user1.LastName, user1.Username, user1.LanguageCode})
	db.Create(&Player{chat.ID, user2.ID, user2.FirstName, user2.LastName, user2.Username, user2.LanguageCode})

	game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "он вышел из этого чата")
}

func TestPlayCommandLaunchGameAndRespondWinner(t *testing.T) {
	defer tearUp(t)()
	db.Create(&Player{chat.ID, user1.ID, user1.FirstName, user1.LastName, user1.Username, user1.LanguageCode})
	db.Create(&Player{chat.ID, user2.ID, user2.FirstName, user2.LastName, user2.Username, user2.LanguageCode})

	game.play(&tb.Message{Text: "/play", Sender: user1, Chat: chat})
	assert.Equal(t, len(messages), 4, "Game response must be multiple")

	var count int64
	db.First(&Entry{}, "chat_id = ?", chat.ID).Count(&count)
	assert.Equal(t, count, int64(1))
}

func TestAllCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	chat := &tb.Chat{Type: tb.ChatPrivate}
	game.all(&tb.Message{Text: "/all", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "недоступна в личных")
}

func TestAllCommandNotRespondsIfNoGamesPlayedYet(t *testing.T) {
	defer tearUp(t)()
	game.all(&tb.Message{Text: "/all", Sender: user1, Chat: chat})
	assert.Equal(t, len(messages), 0)
}

func TestAllCommandRespondsWithAllTimeStat(t *testing.T) {
	defer tearUp(t)()
	day := time.Now().Format("2006-01-02")
	db.Create(&Entry{chat.ID, user1.ID, day, user1.Username})
	game.all(&tb.Message{Text: "/all", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "за всё время")
}

func TestStatsCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	chat := &tb.Chat{Type: tb.ChatPrivate}
	game.stats(&tb.Message{Text: "/stats", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "недоступна в личных")
}

func TestStatsCommandNotRespondingWhenNoGames(t *testing.T) {
	defer tearUp(t)()
	game.stats(&tb.Message{Text: "/stats", Sender: user1, Chat: chat})
	assert.Equal(t, len(messages), 0)
}

func TestStatsCommandRespondsWithCurrentYearStat(t *testing.T) {
	defer tearUp(t)()
	day := time.Now().Format("2006-01-02")
	db.Create(&Entry{chat.ID, user1.ID, day, user1.Username})
	game.stats(&tb.Message{Text: "/stats", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "за текущий год")
}

func TestMeCommandRespondsOnlyInGroupChat(t *testing.T) {
	defer tearUp(t)()
	chat := &tb.Chat{Type: tb.ChatPrivate}
	game.me(&tb.Message{Text: "/me", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "недоступна в личных")
}

func TestMeCommandRespondsWithPersonalStat(t *testing.T) {
	defer tearUp(t)()

	db.Create(&Entry{chat.ID, user1.ID, fmt.Sprintf("%d-01-01", time.Now().Year()), user1.Username})
	db.Create(&Entry{chat.ID, user1.ID, fmt.Sprintf("%d-01-02", time.Now().Year()), user1.Username})
	db.Create(&Entry{chat.ID, user1.ID, fmt.Sprintf("%d-01-03", time.Now().Year()), user1.Username})
	db.Create(&Entry{chat.ID, user1.ID, fmt.Sprintf("%d-01-04", time.Now().Year()), user1.Username})

	game.me(&tb.Message{Text: "/me", Sender: user1, Chat: chat})
	assert.Contains(t, messages[0].Text, "4 раз")
}
