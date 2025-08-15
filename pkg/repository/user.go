package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type UserRepo interface {
	AddUser(user *model.User) error
	FindUser(chatID model.ChatID) (*model.User, error)

	AddMangaToSaved(chatID model.ChatID, manga *model.Manga) error
	RemoveMangaFromSaved(chatID model.ChatID, manga *model.Manga) error
}

type UserRepoSqlite3 struct {
	db *sql.DB
}