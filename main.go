package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	setSLogger()
	err := godotenv.Load()
	if err != nil {
		panic("could not load the .env")
	}
	telegramKey := os.Getenv("TELEGRAM_API_KEY")
	if telegramKey == "" {
		panic("TELEGRAM_API_KEY is not present in the env variables")
	}

	StartTelegramBot(telegramKey)
}



