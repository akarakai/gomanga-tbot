package telegram

import (
	bytes2 "bytes"
	"context"
	"fmt"

	"github.com/akarakai/gomanga-tbot/pkg/downloader"
	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
	"github.com/akarakai/gomanga-tbot/pkg/repository"
	"github.com/akarakai/gomanga-tbot/pkg/scraper"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update, userRepo repository.UserRepo) {
	infoHandler(ctx, b, update)
	registrationHandler(ctx, b, update, userRepo)
}

func registrationHandler(ctx context.Context, b *bot.Bot, update *models.Update, userRepo repository.UserRepo) {
	// save the user in the db if not present
	chatID := model.ChatID(update.Message.Chat.ID)
	usr, err := userRepo.FindUserByChatID(chatID)
	if err != nil {
		logger.Log.Errorw("error when finding user", "err", err)
		sendMessage(ctx, b, int64(chatID), `
There was a problem with the server and you cannot be notified by the bot in the future.
Try again by inserting the command /start again.
		`, nil)
		return
	}

	if usr == nil {
		err := userRepo.SaveUser(chatID)
		if err != nil {
			logger.Log.Errorw("error when saving user", "err", err)
			sendMessage(ctx, b, int64(chatID), `
There was a problem with the server and you cannot be notified by the bot in the future.
Try again by inserting the command /start again.
			`, nil)
			return
		}
		logger.Log.Infow("user saved in the database", "usr_chatID", chatID)
		sendMessage(ctx, b, int64(chatID), "you registered yourself successfully", nil)
		return
	}
	logger.Log.Infow("user already in the database", "usr_chatID", chatID)
	sendMessage(ctx, b, int64(chatID), "you are already registered", nil)
}

func infoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	const welcomeMsg = `Welcome to gomanga-tbot!
Here you can keep track of your favourite mangas published in WeebCentral.
You can also download the latest chapter or read directly on WeebCentral.
Subscribe to a manga, and as soon as it's ready on WeebCentral you will be notified via this bot.

Commands:
/info - Show this help message
/register - Register yourself to get updates. Normally you are automatically registered when you entered the chat (only your chat_id is saved in the server). Call this command if you have problems.
/add <manga name> - Add a manga to your subscription list`
	sendMessage(ctx, b, update.Message.Chat.ID, welcomeMsg, nil)
}

func addHandler(ctx context.Context, b *bot.Bot, update *models.Update, db repository.Database, scraper scraper.Scraper) {
	const cmd = "/add"
	if update.Message == nil {
		logger.Log.Error("Update message is nil")
		return
	}
	if update.Message.From == nil {
		logger.Log.Error("Message.From is nil")
		return
	}
	_ = db.GetMangaRepo()
	_ = db.GetUserRepo()

	userId := update.Message.From.ID
	logger.Log.Infow("new add request", "userId", userId)
	rawMsg := update.Message.Text
	msg, err := parseMessage(cmd, rawMsg)
	if err != nil {
		logger.Log.Debugw("error in message of user", "err", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "to add a manga, use /add 'manga name', without the ''", nil)
		return
	}

	mangas, err := scraper.FindListOfMangas(msg)
	if err != nil {
		logger.Log.Errorw("error creating scraper", "err", err)
		sendMessage(ctx, b, update.Message.Chat.ID, "there are some problems with the bot, try again", nil)
		return
	}

	chatId := model.ChatID(update.Message.Chat.ID)
	convStore.InsertMangas(chatId, mangas)

	// send manga titles as buttons, in the same order of the slices
	keyboard := createMangaKeyboard(mangas)
	sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf("You have chosen: %s", msg), &models.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	})

	logger.Log.Infow("Add message sent successfully", "chatId", update.Message.Chat.ID)
	convStore.InsertAddMangaState(chatId, ChosenManga)
}

// for now it supports only /add
// maybe a more complex arch is needed for supporting conversations
// which start with different commands
func conversationHandler(ctx context.Context, b *bot.Bot, update *models.Update, db repository.Database, scraper scraper.Scraper) {
	logger.Log.Debugln("starting a conversation")
	// get the state from the map
	chatId := model.ChatID(update.Message.Chat.ID)
	state, err := convStore.GetAddMangaState(chatId)
	// handle no chatId saved
	if err != nil {
		logger.Log.Errorln("userId not found in the conversation map")
		return
	}

	switch state {
	case ChosenManga:
		mangaChosenStep(ctx, b, update, db, scraper)
	case ChoseWhatToDo:
		actionOnMangaStep(ctx, b, update)
	default:
		panic("unhandled default case")
	}
}

// second step for /add
// manage the chosen manga from the list
// replies the user with list of actions (download, read online, nothing)
func mangaChosenStep(ctx context.Context, b *bot.Bot, update *models.Update, db repository.Database, scraper scraper.Scraper) {
	logger.Log.Debugf("conversation continues.. Manga was chosen. Now its time for the action")
	chatID := model.ChatID(update.Message.Chat.ID)
	chosenManga := update.Message.Text // sanitation not required
	// get the manga chosen
	mangas, err := convStore.GetMangas(chatID)
	if err != nil {
		logger.Log.Errorf("could not find the manga in the cache")
		sendMessage(ctx, b, int64(chatID), "Could not find the manga in cache", nil)
		return
	}

	var manga model.Manga
	for _, mangaa := range mangas {
		if mangaa.Title == chosenManga {
			logger.Log.Debugw("manga found", "manga", chosenManga)
			manga = mangaa
			break
		}
	}
	if manga == (model.Manga{}) {
		logger.Log.Errorf("manga was not present in the cache")
		sendMessage(ctx, b, int64(chatID), "Manga not found in cache", nil)
		return
	}

	// control if manga is already in the database of the user
	mangaRepo := db.GetMangaRepo()
	if mangas, err := mangaRepo.FindMangasOfUser(chatID); err == nil && len(mangas) > 0 {
		// check if manga is present in the library of the user
		for _, m := range mangas {
			if m.Url == manga.Url {
				logger.Log.Debugw("manga found", "manga", chosenManga)
				// notify user that he is already subscribed
				removeKeyboardFromUser(ctx, b, int64(chatID), "you already are subscribed to this manga")
				convStore.Clean(chatID)
				return
			}
		}
	}

	// user does not have this manga in the repository
	// before scraping the manga, search in the database to avoid open the scraper
	mangaInRepo, err := mangaRepo.FindMangaByUrl(manga.Url)
	if err != nil {
		logger.Log.Errorf("could not find the manga in the repo")
	}
	if mangaInRepo != nil {
		// send to the user the manga of the database
		logger.Log.Debugw("manga found", "manga", chosenManga)
		lastCh := mangaInRepo.LastChapter
		if lastCh == nil { // if db is good written, this should not happen
			logger.Log.Errorln("manga not found in the repo")
			return
		}
		releaseDate := formatReleaseDate(lastCh.ReleasedAt)
		mangaInfoStr := fmt.Sprintf("📚 **%s**\n\n📖 Latest Chapter: %s\n📅 Released: %s\n\nWhat would you like to do?",
			mangaInRepo.Title, lastCh.Title, releaseDate)
		sendMessage(ctx, b, int64(chatID), mangaInfoStr, nil)

		convStore.InsertChosenManga(chatID, *mangaInRepo)
		convStore.InsertAddMangaState(chatID, ChoseWhatToDo)

		keyboard := createActionKeyboard()
		sendMessage(ctx, b, update.Message.Chat.ID, "Please choose an action:", &models.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		})

	}

	chs, err := scraper.FindListOfChapters(manga.Url, 1)
	if err != nil {
		logger.Log.Errorw("error when getting chapters", "err", err)
		sendMessage(ctx, b, int64(chatID), "there was a problem with your bot, try again", nil)
		convStore.Clean(chatID)
		return
	}
	ch := chs[0]
	manga.LastChapter = &ch

	// Format the release date to be more human-readable
	releaseDate := formatReleaseDate(ch.ReleasedAt)

	mangaInfo := fmt.Sprintf("📚 **%s**\n\n📖 Latest Chapter: %s\n📅 Released: %s\n\nWhat would you like to do?",
		manga.Title, ch.Title, releaseDate)

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
	logger.Log.Debugf("conversation continues.. Action was chosen")
	chatID := model.ChatID(update.Message.Chat.ID)
	defer convStore.Clean(chatID)
	choice := CommandManga(update.Message.Text)
	logger.Log.Debugf("user chose: %s", choice)

	manga, err := convStore.GetChosenManga(chatID)
	if err != nil {
		logger.Log.Errorf("manga not found")
		sendMessage(ctx, b, update.Message.Chat.ID, "Manga not found in cache", nil)
		return
	}

	switch choice {
	case Download:
		logger.Log.Infow("user decided to download manga", "manga", manga)
		s, err := scraper.NewWeebCentralScraperDefault()
		if err != nil {
			logger.Log.Errorw("error when creating a scraper", "err", err)
			removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
				"there was a problem when downloading the chapter, try later")
			break
		}
		defer s.Close()

		imgUrls, err := s.FindImgUrlsOfChapter(manga.LastChapter.Url)
		if err != nil {
			logger.Log.Errorw("error when getting chapter imgUrls", "err", err)
			removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
				"there was a problem when downloading the chapter, try later")
			break
		}
		docTitle := fmt.Sprintf("%s-%s", manga.Title, manga.LastChapter.Title)
		bytes, err := downloader.DownloadPdfFromImageSrcs(imgUrls, docTitle)
		if err != nil {
			logger.Log.Errorw("error when constructing the pdf", "err", err)
			removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
				"there was a problem when downloading the chapter, try later")
			break
		}
		logger.Log.Infow("pdf downloaded", "title", docTitle, "sizeBytes", len(bytes))
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
			logger.Log.Errorw("error sending pdf", "err", err)
			return
		}

		logger.Log.Infoln("pdf sent successfully")

	case ReadOnline:
		logger.Log.Infow("user decided to read the manga online", "manga", manga)
		sendMessage(ctx, b, update.Message.Chat.ID, manga.LastChapter.Url, &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		})
		sendMessage(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("You will get a message when the last chapter of %s is released on WeebCentral", manga.Title), nil)
	case DoNothing:
		logger.Log.Infow("user decided to do nothing", "manga", manga)
		removeKeyboardFromUser(ctx, b, update.Message.Chat.ID,
			fmt.Sprintf("You will get a message when the last chapter of %s is released on WeebCentral", manga.Title))
	default:
		removeKeyboardFromUser(ctx, b, update.Message.Chat.ID, "Invalid choice. Please try again with /add command.")
	}
}

// /cancel handler
// cleans the maps from the chatId data
func cancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatId := model.ChatID(update.Message.Chat.ID)
	logger.Log.Infow("deleting conversation history", "chatID", chatId)
	convStore.Clean(chatId)
	removeKeyboardFromUser(ctx, b, update.Message.Chat.ID, "Conversation cancelled. Insert a new command")
}
