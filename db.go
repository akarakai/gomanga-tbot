package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type MangaRepo interface {
	SaveManga(manga *Manga) error
	FindMangasOfChatID(chatID ChatID) ([]Manga, error)	
}

type ChapterRepo interface {
	UpdateLastChapter(chapter *Chapter, mangaUrl string) error
}

type UserRepo interface {
	AddMangaToSaved(manga *Manga) error
	RemoveMangaFromSaved(manga *Manga) error
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
	// which methods?
	Close() error
}

type Sqlite3Database struct {
	db 			*sql.DB
	MangaRepo 	MangaRepo
	ChapterRepo ChapterRepo
	UserRepo 	UserRepo
}

func NewSqlite3Database(dbPath string) (*Sqlite3Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		Log.Errorw("could not connect to the database", "err", err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		Log.Errorw("there was a problem when pinging to the database", "err", err)
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
		db: db,
		MangaRepo: &MangaRepoSqlite3{ db: db },
		ChapterRepo: &ChapterRepoSqlite3{ db: db},
		UserRepo: &UserRepoSqlite3{ db: db},
	}, nil


}


func (repo *MangaRepoSqlite3) SaveManga(manga *Manga) error {
	// TODO
	return nil
}
func (repo *MangaRepoSqlite3) FindMangasOfChatID(chatID ChatID) ([]Manga, error)	{
	// TODO
	return nil, nil
}

func(repo *ChapterRepoSqlite3) UpdateLastChapter(chapter *Chapter, mangaUrl string) error {
	return nil
}

func (repo *UserRepoSqlite3) AddMangaToSaved(manga *Manga) error {
	return nil
}

func (repo *UserRepoSqlite3) RemoveMangaFromSaved(manga *Manga) error {
	return nil
} 
