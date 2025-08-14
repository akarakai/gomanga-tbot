package gomanga

import (
	"os"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	logger.LoggerInit()
	err := godotenv.Load()

	if err != nil {
		logger.Log.Panicw("could not load the .env", "err", err)
	}

	telegramKey := os.Getenv("TELEGRAM_API_KEY")
	if telegramKey == "" {
		logger.Log.Panicln("TELEGRAM_API_KEY is not present in the env variables")
	}

	StartTelegramBot(telegramKey)
}
