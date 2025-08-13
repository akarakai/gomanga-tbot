package main

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// StartTelegramBot starts the bot. Panics if the bot fails to start
func StartTelegramBot(apiKey string) {
	opts := []bot.Option{
		bot.WithDefaultHandler(infoHandler),
	}

	b, err := bot.New(apiKey, opts...)
	if err != nil {
		panic("could not start the bot")
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, infoHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/info", bot.MatchTypeExact, infoHandler)

	b.Start(context.Background())
}

func infoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const welcomeMsg = `Welcome to gomanga-tbot!

Here you can keep track of your favourite mangas published in WeebCentral.
You can also download the latest chapter or read directly on WeebCentral.
Subscribe to a manga, and as soon as it's ready on WeebCentral you will be notified via this bot.

Commands:
`
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   welcomeMsg,
	})
	if err != nil {
		slog.Warn("error sending message", "error", err)		
	}
}