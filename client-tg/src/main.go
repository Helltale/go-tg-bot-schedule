package main

import (
	"fmt"
	"log"
	"tgclient/config"
	"tgclient/internal/handlecommand"
	"tgclient/internal/inlinecommand"

	"github.com/mymmrac/telego"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error: bad token: %s", err)
	}

	bot, err := telego.NewBot(config.Token, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("error: can not start bot: %s", err)
	}

	// inline подсказки
	inlinecommand.New(bot)

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		fmt.Printf("error: error long polling: %s\n", err)
	}
	defer bot.StopLongPolling()

	// handle обработчики
	handlecommand.New(bot, updates)
}
