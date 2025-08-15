package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type MangaRepo interface {
	SaveManga(manga *model.Manga, chatID model.ChatID) error
	FindMangasOfChatID(chatID model.ChatID) ([]model.Manga, error)
}

type MangaRepoSqlite3 struct {
	db *sql.DB
}

// SaveManga saves the manga in the database, along with its last chapter
func (repo *MangaRepoSqlite3) SaveManga(manga *model.Manga, chatID model.ChatID) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	// If there’s a LastChapter, insert it
	if manga.LastChapter != nil {
		_, err = tx.Exec(`
			INSERT OR REPLACE INTO chapters (Url, Title, released_at)
			VALUES (?, ?, ?)`,
			manga.LastChapter.Url,
			manga.LastChapter.Title,
			manga.LastChapter.ReleasedAt,
		)
		if err != nil {
			tx.Rollback()
			logger.Log.Errorw("error when saving chapter", "chapter", manga.LastChapter, "err", err)
			return err
		}
	}

	// Insert manga
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO mangas (Url, Title, last_chapter)
		VALUES (?, ?, ?)`,
		manga.Url,
		manga.Title,
		func() interface{} {
			if manga.LastChapter != nil {
				return manga.LastChapter.Url
			}
			return nil
		}(),
	)
	if err != nil {
		tx.Rollback()
		logger.Log.Errorw("error when saving manga", "manga", manga, "err", err)
		return err
	}

	// add also in the join table chatID with  manga url
	_, err = tx.Exec(`
		INSERT INTO user_mangas (user_id, manga_Url)
		VALUES (?, ?)`,
		chatID,
		manga.Url,
	)
	if err != nil {
		tx.Rollback()
		logger.Log.Errorw("error when saving manga and user in joit table", "manga", manga, "err", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Log.Debugw("manga saved in repo", "manga", manga.Title)
	return nil
}

func (repo *MangaRepoSqlite3) FindMangasOfChatID(chatID model.ChatID) ([]model.Manga, error) {
	rows, err := repo.db.Query(`
		SELECT m.Url, m.Title, c.Url, c.Title, c.released_at
		FROM mangas m
		JOIN user_mangas um ON um.manga_Url = m.Url
		JOIN users u ON u.user_id = um.user_id
		LEFT JOIN chapters c ON m.last_chapter = c.Url
		WHERE u.chat_id = ?`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangas []model.Manga
	for rows.Next() {
		var m model.Manga
		var chURL, chTitle sql.NullString
		var chReleased sql.NullTime

		if err := rows.Scan(&m.Url, &m.Title, &chURL, &chTitle, &chReleased); err != nil {
			return nil, err
		}

		if chURL.Valid {
			m.LastChapter = &model.Chapter{
				Url:        chURL.String,
				Title:      chTitle.String,
				ReleasedAt: chReleased.Time,
			}
		}
		mangas = append(mangas, m)
	}
	return mangas, nil
}