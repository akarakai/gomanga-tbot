package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

type MangaRepo interface {
	SaveManga(manga *main.Manga) error
	FindMangasOfChatID(chatID main.ChatID) ([]main.Manga, error)
}

type ChapterRepo interface {
	UpdateLastChapter(chapter *main.Chapter, mangaUrl string) error
}

type UserRepo interface {
	AddMangaToSaved(manga *main.Manga) error
	RemoveMangaFromSaved(manga *main.Manga) error
}

type MangaRepoSqlite3 struct {
	db *sql.DB
}
type ChapterRepoSqlite3 struct {
	db *sql.DB
}
type UserRepoSqlite3 struct {
	db *sql.DB
}

type Database interface {
	Close() error
}

type Sqlite3Database struct {
	db          *sql.DB
	MangaRepo   MangaRepo
	ChapterRepo ChapterRepo
	UserRepo    UserRepo
}

func NewSqlite3Database(dbPath string) (*Sqlite3Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Log.Errorw("could not connect to the database", "err", err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		logger.Log.Errorw("there was a problem when pinging to the database", "err", err)
		return nil, err
	}

	// create the tables
	// 	// Enable foreign keys
	db.Exec(`PRAGMA foreign_keys = ON;`)

	// Create chapters table
	db.Exec(`
	CREATE TABLE IF NOT EXISTS chapters (
		url TEXT PRIMARY KEY,
		title TEXT NOT NULL UNIQUE,
		released_at DATETIME NOT NULL
	);`)

	// Create mangas table
	db.Exec(`
	CREATE TABLE IF NOT EXISTS mangas (
		url TEXT PRIMARY KEY,
		title TEXT NOT NULL UNIQUE,
		last_chapter TEXT,
		FOREIGN KEY (last_chapter) REFERENCES chapters(url) ON DELETE SET NULL
	);`)

	// Create users table
	db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY,
		chat_id INTEGER NOT NULL
	);`)

	// Create user_mangas join table (many-to-many)
	db.Exec(`
	CREATE TABLE IF NOT EXISTS user_mangas (
		user_id INTEGER NOT NULL,
		manga_url TEXT NOT NULL,
		PRIMARY KEY (user_id, manga_url),
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
		FOREIGN KEY (manga_url) REFERENCES mangas(url) ON DELETE CASCADE
	);`)

	return &Sqlite3Database{
		db:          db,
		MangaRepo:   &MangaRepoSqlite3{db: db},
		ChapterRepo: &ChapterRepoSqlite3{db: db},
		UserRepo:    &UserRepoSqlite3{db: db},
	}, nil

}

func (repo *MangaRepoSqlite3) SaveManga(manga *main.Manga) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	// If there’s a lastChapter, insert it
	if manga.lastChapter != nil {
		_, err = tx.Exec(`
			INSERT OR REPLACE INTO chapters (url, title, released_at)
			VALUES (?, ?, ?)`,
			manga.lastChapter.url,
			manga.lastChapter.title,
			manga.lastChapter.releasedAt,
		)
		if err != nil {
			tx.Rollback()
			logger.Log.Errorw("error when saving chapter", "chapter", manga.lastChapter, "err", err)
			return err
		}
	}

	// Insert manga
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO mangas (url, title, last_chapter)
		VALUES (?, ?, ?)`,
		manga.url,
		manga.title,
		func() interface{} {
			if manga.lastChapter != nil {
				return manga.lastChapter.url
			}
			return nil
		}(),
	)
	if err != nil {
		tx.Rollback()
		logger.Log.Errorw("error when saving manga", "manga", manga, "err", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Log.Debugw("manga saved in repo", "manga", manga.title)
	return nil
}

func (repo *MangaRepoSqlite3) FindMangasOfChatID(chatID main.ChatID) ([]main.Manga, error) {
	rows, err := repo.db.Query(`
		SELECT m.url, m.title, c.url, c.title, c.released_at
		FROM mangas m
		JOIN user_mangas um ON um.manga_url = m.url
		JOIN users u ON u.user_id = um.user_id
		LEFT JOIN chapters c ON m.last_chapter = c.url
		WHERE u.chat_id = ?`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangas []main.Manga
	for rows.Next() {
		var m main.Manga
		var chURL, chTitle sql.NullString
		var chReleased sql.NullTime

		if err := rows.Scan(&m.url, &m.title, &chURL, &chTitle, &chReleased); err != nil {
			return nil, err
		}

		if chURL.Valid {
			m.lastChapter = &main.Chapter{
				url:        chURL.String,
				title:      chTitle.String,
				releasedAt: chReleased.Time,
			}
		}
		mangas = append(mangas, m)
	}
	return mangas, nil
}

func (repo *ChapterRepoSqlite3) UpdateLastChapter(chapter *main.Chapter, mangaUrl string) error {
	_, err := repo.db.Exec(`
		INSERT OR REPLACE INTO chapters (url, title, released_at)
		VALUES (?, ?, ?)`,
		chapter.url, chapter.title, chapter.releasedAt)
	if err != nil {
		return err
	}

	_, err = repo.db.Exec(`
		UPDATE mangas SET last_chapter = ? WHERE url = ?`,
		chapter.url, mangaUrl)
	return err
}

func (repo *UserRepoSqlite3) AddMangaToSaved(manga *main.Manga) error {
	_, err := repo.db.Exec(`
		INSERT OR IGNORE INTO mangas (url, title) VALUES (?, ?)`,
		manga.url, manga.title)
	if err != nil {
		return err
	}

	// Dummy: assuming user_id = 1 for now
	_, err = repo.db.Exec(`
		INSERT OR IGNORE INTO user_mangas (user_id, manga_url)
		VALUES (?, ?)`, 1, manga.url)
	return err
}

func (repo *UserRepoSqlite3) RemoveMangaFromSaved(manga *main.Manga) error {
	_, err := repo.db.Exec(`
		DELETE FROM user_mangas WHERE manga_url = ?`, manga.url)
	return err
}

func (s *Sqlite3Database) Close() error {
	return s.db.Close()
}
