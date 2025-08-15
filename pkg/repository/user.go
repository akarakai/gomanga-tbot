package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type UserRepo interface {
	SaveUser(chatID model.ChatID) error
	RegisterToManga(chatID model.ChatID, mangaUrl string) error
	FindUserByChatID(chatID model.ChatID) (*model.User, error)
}

type UserRepoSqlite3 struct {
	db *sql.DB
}

func (repo *UserRepoSqlite3) SaveUser(chatID model.ChatID) error {
	_, err := repo.db.Exec(`
		INSERT OR IGNORE INTO users (chat_id)
		VALUES (?)
	`, chatID)
	if err != nil {
		logger.Log.Errorw("error when saving user", "chat_id", chatID, "err", err)
		return err
	}

	logger.Log.Debugw("user saved successfully", "chat_id", chatID)
	return nil
}

func (repo *UserRepoSqlite3) FindUserByChatID(chatID model.ChatID) (*model.User, error) {
	row := repo.db.QueryRow(`
        SELECT chat_id FROM users
        WHERE chat_id = ?
    `, chatID)

	var chatIDq sql.NullInt64
	if err := row.Scan(&chatIDq); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debugw("user does not exist", "chat_id", chatIDq.Int64)
			return nil, nil
		}
		logger.Log.Errorw("error when scanning user row", "err", err)
		return nil, err
	}

	logger.Log.Debugw("user found successfully", "chat_id", chatIDq.Int64)

	return &model.User{
		ChatID: model.ChatID(chatIDq.Int64),
	}, nil
}

func (repo *UserRepoSqlite3) RegisterToManga(chatID model.ChatID, mangaURL string) error {
	_, err := repo.db.Exec(`
		INSERT INTO user_mangas (chat_id, manga_url)
		VALUES (?, ?)
		ON CONFLICT(chat_id, manga_url) DO NOTHING
	`, chatID, mangaURL)
	if err != nil {
		logger.Log.Errorw("register user to manga failed",
			"chat_id", chatID, "manga_url", mangaURL, "err", err)
		return err
	}
	logger.Log.Debugw("manga saved to user succesfully", "chat_id", chatID)
	return nil
}
