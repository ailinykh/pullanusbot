package main

import (
	"os"
	"testing"
)

func TestMainShouldReturnIfNoBotToken(t *testing.T) {
	main()
	// t.Error("main() should return if no BOT_TOKEN passed")
}

func TestMainShouldReturnIfTokenWrong(t *testing.T) {
	os.Setenv("BOT_TOKEN", "some bot token")
	main()
	// t.Error("main() should return if no BOT_TOKEN is wrong")
}
