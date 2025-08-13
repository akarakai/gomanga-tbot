package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	LoggerInit()
	err := godotenv.Load()

	if err != nil {
		Log.Panicw("could not load the .env", "err", err)
	}

	telegramKey := os.Getenv("TELEGRAM_API_KEY")
	if telegramKey == "" {
		Log.Panicln("TELEGRAM_API_KEY is not present in the env variables")
	}

	StartTelegramBot(telegramKey)
}
