package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var convStore = NewConversationStore()

// StartTelegramBot starts the bot. Panics if the bot fails to start
func StartTelegramBot(apiKey string) {
	b, err := bot.New(apiKey)
	if err != nil {
		Log.Panicw("could not start the bot", "err", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "info", bot.MatchTypeCommand, infoHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "add", bot.MatchTypeCommand, addHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "cancel", bot.MatchTypeCommand, cancelHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains, conversationHandler)

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
	msg, err := parseMessage(cmd, rawMsg)
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
			Text:   "there are some problems with the bot, try again",
		})
	}

	chatId := ChatID(update.Message.Chat.ID)
	convStore.InsertMangas(chatId, mangas)

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

	removeKeyboardFromUser(ctx, b, "", update.Message.Chat.ID)

	if err != nil {
		Log.Errorw("Error sending add message", "error", err, "chatId", update.Message.Chat.ID)
		return
	}
	Log.Infow("Add message sent successfully", "chatId", update.Message.Chat.ID)
	convStore.InsertAddMangaState(chatId, ChosenManga)
}

// for now it supports only /add
// maybe a more complex arch is needed for supporting conversations
// which start with different commands
func conversationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	Log.Debugln("starting a conversation")
	// get the state from the map
	chatId := ChatID(update.Message.Chat.ID)
	state, err := convStore.GetAddMangaState(chatId)
	// handle no chatId saved
	if err != nil {
		Log.Errorln("userId not found in the concersation map")
		return
	}

	switch state {
	case ChosenManga:
		mangaChosenStep(ctx, b, update)
	case ChoseWhatToDo:
		actionOnMangaStep(ctx, b, update)
	}
}

// second step for /add
// manage the chosen manga from the list
// replies the user with list of actions (download, read online, nothing)
func mangaChosenStep(ctx context.Context, b *bot.Bot, update *models.Update) {
	Log.Debugf("conversation continues.. Manga was chosen. Now its time for the action")
	chatID := ChatID(update.Message.Chat.ID)
	chosenManga := update.Message.Text // sanitation not required
	// get the manga chosen
	mangas, err := convStore.GetMangas(chatID)
	if err != nil {
		Log.Errorf("could not find the manga in the cache")
		// TODO add message
		return
	}

	var manga Manga
	for _, mangaa := range mangas {
		if mangaa.title == chosenManga {
			Log.Debugw("manga found", "manga", chosenManga)
			manga = mangaa
		}
	}
	if manga == (Manga{}) {
		Log.Errorf("manga was not present in the cache")
		return
	}

	convStore.InsertChosenManga(chatID, manga)
	convStore.InsertAddMangaState(chatID, ChoseWhatToDo)

	keyboard := [][]models.KeyboardButton{
		{{Text: string(Download)}},
		{{Text: string(ReadOnline)}},
		{{Text: string(DoNothing)}},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Please choose an action",
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})

	// remove keyboard choices
	removeKeyboardFromUser(ctx, b, "", update.Message.Chat.ID)
}

// final step for /add
// user chooses what to do with the last manga
func actionOnMangaStep(ctx context.Context, b *bot.Bot, update *models.Update) {
	Log.Debugf("conversation continues.. Action was chosen")
	chatID := ChatID(update.Message.Chat.ID)
	choice := CommandManga(update.Message.Text)
	Log.Debugf("user chose: %s", choice)

	manga, err := convStore.GetChosenManga(chatID)
	if err != nil {
		Log.Errorf("manga not found")
		return
	}

	switch choice {
	case Download:
		Log.Infow("user decided to download manga", "manga", manga)
	case ReadOnline:
		Log.Infow("user decided to read the manga online", "manga", manga)
		url := manga.url
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   url,
		})
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("you will get a message when the last chapter of %s is released on WeebCentral", manga.title),
		})
	case DoNothing:
		Log.Infow("user decided to do nothing", "manga", manga)
		removeKeyboardFromUser(ctx,
			b,
			fmt.Sprintf("you will get a message when the last chapter of %s is released on WeebCentral", manga.title),
			update.Message.Chat.ID)
	}

	// cleanup
	convStore.Clean(chatID)
}

// /cancel handler
// cleans the maps from the chatId data
func cancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatId := ChatID(update.Message.Chat.ID)
	Log.Infow("deleting conversation history", "chatID", chatId)
	convStore.Clean(chatId)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "conversation cancelled. Insert a new command",
	})
}

func removeKeyboardFromUser(ctx context.Context, b *bot.Bot, text string, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
		ReplyMarkup: &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
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
