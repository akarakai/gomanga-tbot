package telegram

import (
	"context"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/repository"
	"github.com/akarakai/gomanga-tbot/pkg/scraper"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Service struct {
	bot     *bot.Bot
	db      repository.Database
	scraper scraper.Scraper
	ctx     context.Context
}

func NewTelegramService(ctx context.Context, apiKey string, db repository.Database, scraper scraper.Scraper) (*Service, error) {
	b, err := bot.New(apiKey)
	if err != nil {
		return nil, err
	}
	return &Service{
		bot:     b,
		db:      db,
		scraper: scraper,
		ctx:     ctx,
	}, nil
}

func (t *Service) Start() {

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand,
		func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			startHandler(ctx, bot, update, t.db.GetUserRepo())
		})

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "list", bot.MatchTypeCommand,
		func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			mangaListHandler(ctx, bot, update, t.db.GetMangaRepo())
		})

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "register", bot.MatchTypeCommand,
		func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			registrationHandler(ctx, bot, update, t.db.GetUserRepo())
		})

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "info", bot.MatchTypeCommand, infoHandler)

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommand, infoHandler)

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "add", bot.MatchTypeCommand,
		func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			addHandler(ctx, bot, update, t.db, t.scraper)

		})

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "cancel", bot.MatchTypeCommand, cancelHandler)

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains,
		func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			conversationHandler(ctx, bot, update, t.db, t.scraper)
		})

	logger.Log.Infof("starting the bot")
	t.bot.Start(t.ctx)
}
