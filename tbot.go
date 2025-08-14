package main

import (
	bytes2 "bytes"
	"context"
	"fmt"
	"strings"
	"time"

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

	sendMessage(ctx, b, update.Message.Chat.ID, welcomeMsg, nil)
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
		sendMessage(ctx, b, update.Message.Chat.ID, "to add a manga, use /add 'manga name', without the ''", nil)
		return
	}

	// search the manga
	s, err := NewWeebCentralScraperDefault()
	if err != nil {
		Log.Errorw("error when creating a scraper", "err", err)
		sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf("You have chosen: %s", msg), nil)
		return
	}

	mangas, err := s.FindListOfMangas(msg)
	if err != nil {
		Log.Errorw("error creating scraper", "err", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "there are some problems with the bot, try again", nil)
		return
	}

	chatId := ChatID(update.Message.Chat.ID)
	convStore.InsertMangas(chatId, mangas)

	// send manga titles as buttons, in the same order of the slices
	keyboard := createMangaKeyboard(mangas)
	sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf("You have chosen: %s", msg), &models.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	})

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
		Log.Errorln("userId not found in the conversation map")
		return
	}

	switch state {
	case ChosenManga:
		mangaChosenStep(ctx, b, update)
	case ChoseWhatToDo:
		actionOnMangaStep(ctx, b, update)
	default:
		panic("unhandled default case")
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
		sendMessage(ctx, b, int64(chatID), "Could not find the manga in cache", nil)
		return
	}

	var manga Manga
	for _, mangaa := range mangas {
		if mangaa.title == chosenManga {
			Log.Debugw("manga found", "manga", chosenManga)
			manga = mangaa
			break
		}
	}
	if manga == (Manga{}) {
		Log.Errorf("manga was not present in the cache")
		sendMessage(ctx, b, int64(chatID), "Manga not found in cache", nil)
		return
	}

	// get last chapter of the manga
	s, err := NewWeebCentralScraperDefault()
	if err != nil {
		Log.Errorw("error when creating a scraper", "err", err)
		sendMessage(ctx, b, int64(chatID), "there was a problem with your bot, try again", nil)
		convStore.Clean(chatID)
		return
	}
	defer s.Close()

	chs, err := s.FindListOfChapters(manga.url, 1)
	if err != nil {
		Log.Errorw("error when getting chapters", "err", err)
		sendMessage(ctx, b, int64(chatID), "there was a problem with your bot, try again", nil)
		convStore.Clean(chatID)
		return
	}
	ch := chs[0]
	manga.lastChapter = &ch

	// Format the release date to be more human-readable
	releaseDate := formatReleaseDate(ch.releasedAt)

	mangaInfo := fmt.Sprintf("📚 **%s**\n\n📖 Latest Chapter: %s\n📅 Released: %s\n\nWhat would you like to do?",
		manga.title, ch.title, releaseDate)

	// First remove the previous keyboard and send manga info
	sendMessage(ctx, b, int64(chatID), mangaInfo, &models.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	})

	convStore.InsertChosenManga(chatID, manga)
	convStore.InsertAddMangaState(chatID, ChoseWhatToDo)

	// Then send the action keyboard
	keyboard := createActionKeyboard()
	sendMessage(ctx, b, update.Message.Chat.ID, "Please choose an action:", &models.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	})
}

// final step for /add
// user chooses what to do with the last manga
func actionOnMangaStep(ctx context.Context, b *bot.Bot, update *models.Update) {
	Log.Debugf("conversation continues.. Action was chosen")
	chatID := ChatID(update.Message.Chat.ID)
	defer convStore.Clean(chatID)
	choice := CommandManga(update.Message.Text)
	Log.Debugf("user chose: %s", choice)

	manga, err := convStore.GetChosenManga(chatID)
	if err != nil {
		Log.Errorf("manga not found")
		sendMessage(ctx, b, update.Message.Chat.ID, "Manga not found in cache", nil)
		return
	}

	switch choice {
	case Download:
		Log.Infow("user decided to download manga", "manga", manga)
		s, err := NewWeebCentralScraperDefault()
		if err != nil {
			Log.Errorw("error when creating a scraper", "err", err)
			removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
				"there was a problem when downloading the chapter, try later")
			break
		}
		defer s.Close()

		imgUrls, err := s.FindImgUrlsOfChapter(manga.lastChapter.url)
		if err != nil {
			Log.Errorw("error when getting chapter imgUrls", "err", err)
			removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
				"there was a problem when downloading the chapter, try later")
			break
		}
		docTitle := fmt.Sprintf("%s-%s", manga.title, manga.lastChapter.title)
		bytes, err := DownloadPdfFromImageSrcs(imgUrls, docTitle)
		if err != nil {
			Log.Errorw("error when constructing the pdf", "err", err)
			removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
				"there was a problem when downloading the chapter, try later")
			break
		}
		Log.Infow("pdf downloaded", "title", docTitle, "sizeBytes", len(bytes))
		// send as pdf
		fileReader := bytes2.NewReader(bytes)

		_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
			ChatID: update.Message.Chat.ID,
			Document: &models.InputFileUpload{
				Filename: fmt.Sprintf("%s.pdf", docTitle),
				Data:     fileReader,
			},
		})
		if err != nil {
			Log.Errorw("error sending pdf", "err", err)
			return
		}

		Log.Infoln("pdf sent successfully")

	case ReadOnline:
		Log.Infow("user decided to read the manga online", "manga", manga)
		sendMessage(ctx, b, update.Message.Chat.ID, manga.lastChapter.url, &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		})
		sendMessage(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("You will get a message when the last chapter of %s is released on WeebCentral", manga.title), nil)
	case DoNothing:
		Log.Infow("user decided to do nothing", "manga", manga)
		removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("You will get a message when the last chapter of %s is released on WeebCentral", manga.title))
	default:
		removeKeyboardFromUser(ctx, b, update.Message.Chat.ID, "Invalid choice. Please try again with /add command.")
	}
}

// /cancel handler
// cleans the maps from the chatId data
func cancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatId := ChatID(update.Message.Chat.ID)
	Log.Infow("deleting conversation history", "chatID", chatId)
	convStore.Clean(chatId)
	removeKeyboardFromUser(ctx, b, update.Message.Chat.ID, "Conversation cancelled. Insert a new command")
}

// ========== HELPER FUNCTIONS ==========

// Helper function to format release date to human-readable format
func formatReleaseDate(releaseTime time.Time) string {
	now := time.Now()
	diff := now.Sub(releaseTime)

	// Format relative time
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes < 1 {
			return "Just now"
		}
		return fmt.Sprintf("%d minute%s ago", minutes, pluralS(minutes))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralS(hours))
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralS(days))
	} else {
		// For older dates, show the actual date
		return releaseTime.Format("January 2, 2006")
	}
}

// Helper function to add 's' for plural
func pluralS(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// Helper function to reduce code duplication for sending messages
func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string, replyMarkup models.ReplyMarkup) {
	if text == "" {
		Log.Warn("Attempting to send empty message, skipping")
		return
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: replyMarkup,
	})
	if err != nil {
		Log.Errorw("Error sending message", "error", err, "chatId", chatID)
	}
}

// Helper function to remove keyboard and send a message
func removeKeyboardFromUser(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	if message == "" {
		message = "Keyboard removed"
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
		ReplyMarkup: &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
	if err != nil {
		Log.Errorw("Error removing keyboard", "error", err, "chatId", chatID)
	}
}

// Helper function to create manga keyboard
func createMangaKeyboard(mangas []Manga) [][]models.KeyboardButton {
	var keyboard [][]models.KeyboardButton
	for _, manga := range mangas {
		row := []models.KeyboardButton{
			{Text: manga.title},
		}
		keyboard = append(keyboard, row)
	}
	return keyboard
}

// Helper function to create action keyboard
func createActionKeyboard() [][]models.KeyboardButton {
	return [][]models.KeyboardButton{
		{{Text: string(Download)}},
		{{Text: string(ReadOnline)}},
		{{Text: string(DoNothing)}},
	}
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
