package faggot

import "testing"

func TestStatistics(t *testing.T) {

	stats := Statistics{}

	stats.Increment("player1")
	stats.Increment("player1")
	stats.Increment("player2")

	// Test Less
	if !stats.Less(1, 0) {
		t.Error("Statistics Less function incorrect behaviour")
	}

	// Test Swap
	stats.Swap(0, 1)

	if stats.stat[0].Player != "player2" {
		t.Error("Statistics Swap function incorrect behaviour")
	}
}
