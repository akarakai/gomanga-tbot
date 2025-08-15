package repository

import (
	"database/sql"
	"errors"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type MangaRepo interface {
	SaveManga(manga *model.Manga, chatID model.ChatID) error
	FindMangaByUrl(url string) (*model.Manga, error)
	FindMangasOfUser(chatID model.ChatID) ([]model.Manga, error)
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
			INSERT OR REPLACE INTO chapters (url, title, released_at)
			VALUES (?, ?, ?)`,
			manga.LastChapter.Url,
			manga.LastChapter.Title,
			manga.LastChapter.ReleasedAt,
		)
		if err != nil {
			_ = tx.Rollback()
			logger.Log.Errorw("error when saving chapter", "chapter", manga.LastChapter, "err", err)
			return err
		}
	}

	// Insert manga
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO mangas (url, title, last_chapter)
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
		_ = tx.Rollback()
		logger.Log.Errorw("error when saving manga", "manga", manga, "err", err)
		return err
	}

	// add also in the join table chatID with  manga url
	_, err = tx.Exec(`
		INSERT INTO user_mangas (chat_id, manga_url)
		VALUES (?, ?)`,
		chatID,
		manga.Url,
	)
	if err != nil {
		_ = tx.Rollback()
		logger.Log.Errorw("error when saving manga and user in joit table", "manga", manga, "err", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Log.Debugw("manga saved in repo", "manga", manga.Title)
	return nil
}

func (repo *MangaRepoSqlite3) FindMangasOfUser(chatID model.ChatID) ([]model.Manga, error) {
	rows, err := repo.db.Query(`
		SELECT m.url, m.title, c.url, c.title, c.released_at
		FROM mangas m
		JOIN user_mangas um ON um.manga_url = m.url
		JOIN users u ON u.chat_id = um.chat_id
		LEFT JOIN chapters c ON m.last_chapter = c.url
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

func (repo *MangaRepoSqlite3) FindMangaByUrl(url string) (*model.Manga, error) {
	row, err := repo.db.Query(`
		SELECT m.url, m.title, c.url, title, c.released_at
		FROM mangas m
		JOIN chapters c ON m.last_chapter = c.url
		WHERE c.url = ?
`, url)
	if err != nil {
		logger.Log.Errorw("error when finding manga by url", "url", url, "err", err)
		return nil, errors.New("error getting manga by url")
	}
	defer func(row *sql.Rows) {
		err := row.Close()
		if err != nil {
			logger.Log.Errorw("error when closing rows", "err", err)
		}
	}(row)

	var mangaURL sql.NullString
	var mangaTitle sql.NullString

	var chapterURL sql.NullString
	var chapterTitle sql.NullString
	var chapterReleased sql.NullTime

	if err := row.Scan(&mangaURL, &mangaTitle, &chapterURL, &chapterTitle, &chapterReleased); err != nil {
		logger.Log.Errorw("error when scanning manga row", "err", err)
		return nil, err
	}

	logger.Log.Debugw("manga found successfully by url", "mangaTitle", mangaTitle, "lastCh", chapterTitle)
	return &model.Manga{
		Title: mangaTitle.String,
		Url:   mangaURL.String,
		LastChapter: &model.Chapter{
			Title:      chapterTitle.String,
			Url:        chapterURL.String,
			ReleasedAt: chapterReleased.Time,
		},
	}, nil

}
