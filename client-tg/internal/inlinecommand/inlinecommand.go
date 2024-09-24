package inlinecommand

import (
	"log"

	"github.com/mymmrac/telego"
)

func New(bot *telego.Bot) {
	commands := []telego.BotCommand{
		{Command: "start", Description: "Перезапустить бота"},
		// {Command: "help", Description: "Получить помощь"},
		// {Command: "settings", Description: "Настройки бота"},
	}

	// Создайте параметры для установки команд
	params := &telego.SetMyCommandsParams{
		Commands: commands,
	}

	err := bot.SetMyCommands(params)
	if err != nil {
		log.Fatalf("Ошибка при установке команд: %v", err)
	}

	log.Println("Команды успешно установлены!")
}
