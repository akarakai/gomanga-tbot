package main

import (
	"context"
	"os"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/repository"
	"github.com/akarakai/gomanga-tbot/pkg/scraper"
	"github.com/akarakai/gomanga-tbot/pkg/telegram"
	"github.com/joho/godotenv"
)

func main() {
	logger.LoggerInit()
	logger.Log.Info("Starting Gomanga Bot")

	err := godotenv.Load()

	if err != nil {
		logger.Log.Errorw("could not load the .env", "err", err)
	}

	telegramKey := os.Getenv("TELEGRAM_API_KEY")
	if telegramKey == "" {
		logger.Log.Panicln("TELEGRAM_API_KEY is not present in the env variables")
	}

	repo, err := repository.NewSqlite3Database("./database.db")
	if err != nil {
		logger.Log.Panicw("could not connect to the database", "err", err)
	}

	s, err := scraper.NewWeebCentralScraperDefault()
	if err != nil {
		logger.Log.Panicw("could not connect to the scraper", "err", err)
	}

	tg, err := telegram.NewTelegramService(
		telegramKey,
		repo,
		s,
	)

	if err != nil {
		logger.Log.Panicw("could not create bot instance", "err", err)
	}

	tg.Start(context.Background())
}
