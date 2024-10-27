package infrastructure

import (
	"math/rand"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateMathRand() core.IRand {
	return MathRand{}
}

type MathRand struct{}

func (MathRand) GetRand(n int) int {
	return rand.Intn(n)
}
