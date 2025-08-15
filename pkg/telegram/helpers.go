package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

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
		logger.Log.Warn("Attempting to send empty message, skipping")
		return
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: replyMarkup,
	})
	if err != nil {
		logger.Log.Errorw("Error sending message", "error", err, "chatId", chatID)
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
		logger.Log.Errorw("Error removing keyboard", "error", err, "chatId", chatID)
	}
}

// Helper function to create manga keyboard
func createMangaKeyboard(mangas []model.Manga) [][]models.KeyboardButton {
	var keyboard [][]models.KeyboardButton
	for _, manga := range mangas {
		row := []models.KeyboardButton{
			{Text: manga.Title},
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
