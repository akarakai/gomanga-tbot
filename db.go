package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite3Repo struct {
	db			*sql.DB		// can do general things
	MangaRepo 	*MangaRepo
	ChapterRepo *ChapterRepo
	UserRepo 	*UserRepo
}

type MangaRepo struct {
	db *sql.DB
}
type ChapterRepo struct {
	db *sql.DB
}
type UserRepo struct {
	db *sql.DB
}

func NewSqlite3Repo(databasePath string) *Sqlite3Repo {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		Log.Panicw("panic creating database", "err", err)
	}

	// Enable foreign keys
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
		last_chapter TEXT NOT NULL,
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

	return &Sqlite3Repo {
		db: db,
		MangaRepo: &MangaRepo{ db: db },
		ChapterRepo: &ChapterRepo{ db: db },
		UserRepo: &UserRepo{ db: db },
	}
} 
