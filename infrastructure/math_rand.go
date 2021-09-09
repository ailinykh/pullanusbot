package infrastructure

import (
	"math/rand"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateMathRand() core.IRand {
	rand.Seed(time.Now().UTC().UnixNano())
	return MathRand{}
}

type MathRand struct{}

func (MathRand) GetRand(n int) int {
	return rand.Intn(n)
}
