package infrastructure

import (
	c "crypto/rand"
	"encoding/binary"
	"math/rand"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateCryptoRand() core.IRand {
	var src cryptoSeed
	return CryptoRand{rand.New(src)}
}

type CryptoRand struct {
	r *rand.Rand
}

func (c CryptoRand) GetRand(n int) int {
	return c.r.Intn(n)
}

type cryptoSeed struct{}

func (cryptoSeed) Seed(seed int64) {}

func (r cryptoSeed) Int63() int64 {
	return int64(r.Uint64() & ^uint64(1<<63))
}

func (cryptoSeed) Uint64() (v uint64) {
	err := binary.Read(c.Reader, binary.BigEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}
