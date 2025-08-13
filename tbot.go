package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// StartTelegramBot starts the bot. Panics if the bot fails to start
func StartTelegramBot(apiKey string) {
	opt := []bot.Option{
		bot.WithDefaultHandler(infoHandler),
	}

	b, err := bot.New(apiKey, opt...)
	if err != nil {
		Log.Panicw("could not start the bot", "err", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "info", bot.MatchTypeCommand, infoHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "add", bot.MatchTypeCommand, addHandler)

	Log.Infof("starting the bot")
	b.Start(context.Background())
}

func infoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const welcomeMsg = `Welcome to gomanga-tbot!
Here you can keep track of your favourite mangas published in WeebCentral.
You can also download the latest chapter or read directly on WeebCentral.
Subscribe to a manga, and as soon as it's ready on WeebCentral you will be notified via this bot.

Commands:
/info - Show this help message
/add <manga name> - Add a manga to your subscription list`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   welcomeMsg,
	})
	if err != nil {
		Log.Warnw("error sending message", "err", err)
	}
}

func addHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const cmd = "/add"
	if update.Message == nil {
		Log.Error("Update message is nil")
		return
	}

	if update.Message.From == nil {
		Log.Error("Message.From is nil")
		return
	}

	userId := update.Message.From.ID

	Log.Infow("new add request", "userId", userId)

	rawMsg := update.Message.Text
	msg, err := parseMessage("/add", rawMsg)
	if err != nil {
		Log.Debugw("error in message of user", "err", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "to add a manga, use \\add 'manga name', without the ''",
		})
		if err != nil {
			Log.Error("Error sending add message", "error", err, "chatId", update.Message.Chat.ID)
			return
		}
	}

	// search the manga
	s, err := NewWeebCentralScraperDefault()
	if err != nil {
		Log.Errorw("error when creating a scraper", "err", err)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("You have chosen: %s", msg),
		})
		if err != nil {
			Log.Error("Error sending add message", "error", err, "chatId", update.Message.Chat.ID)
			return
		}

	}

	mangas, err := s.FindListOfMangas(msg)
	if err != nil {
		Log.Errorw("error creating scraper", "err", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("there are some problems with the bot, try again"),
		})
	}

	// send manga titles as buttons, in the same order of the slices
	var keyboard [][]models.KeyboardButton
	for _, manga := range mangas {
		row := []models.KeyboardButton{
			{Text: manga.title},
		}
		keyboard = append(keyboard, row)
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("You have chosen: %s", msg),
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})
	if err != nil {
		Log.Errorw("Error sending add message", "error", err, "chatId", update.Message.Chat.ID)
		return
	}
	Log.Infow("Add message sent successfully", "chatId", update.Message.Chat.ID)
}

// /add One Piece => One Piece
func parseMessage(command string, fullMessage string) (string, error) {
	if !strings.HasPrefix(fullMessage, command) {
		return "", fmt.Errorf("not a command %s", fullMessage)
	}

	splits := strings.Split(fullMessage, " ")
	// remove command. Note this is a single word command
	splits = splits[1:]
	msg := strings.Join(splits, " ")
	trimmed := strings.Trim(msg, " \n\t")
	return trimmed, nil
}
