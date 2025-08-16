package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

type Database interface {
	GetMangaRepo() MangaRepo
	GetUserRepo() UserRepo
	GetChapterRepo() ChapterRepo
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

	loadTables(db)

	return &Sqlite3Database{
		db:          db,
		MangaRepo:   &MangaRepoSqlite3{db: db},
		ChapterRepo: &ChapterRepoSqlite3{db: db},
		UserRepo:    &UserRepoSqlite3{db: db},
	}, nil

}

func (s *Sqlite3Database) Close() error {
	return s.db.Close()
}

func (s *Sqlite3Database) GetChapterRepo() ChapterRepo {
	if s.ChapterRepo == nil {
		logger.Log.Panicln("chapter repo not initialized")
	}
	return s.ChapterRepo
}

func (s *Sqlite3Database) GetUserRepo() UserRepo {
	if s.UserRepo == nil {
		logger.Log.Panicln("chapter repo not initialized")
	}
	return s.UserRepo
}

func (s *Sqlite3Database) GetMangaRepo() MangaRepo {
	if s.MangaRepo == nil {
		logger.Log.Panicln("chapter repo not initialized")
	}
	return s.MangaRepo
}

func loadTables(db *sql.DB) {
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
			chat_id INTEGER NOT NULL PRIMARY KEY
		);`)

		// Create user_mangas join table (many-to-many)
		db.Exec(`
		CREATE TABLE IF NOT EXISTS user_mangas (
			chat_id INTEGER NOT NULL,
			manga_url TEXT NOT NULL,
			PRIMARY KEY (chat_id, manga_url),
			FOREIGN KEY (chat_id) REFERENCES users(chat_id) ON DELETE CASCADE,
			FOREIGN KEY (manga_url) REFERENCES mangas(url) ON DELETE CASCADE
		);`)
	}

// func removeDatabaseTestFile() error {
// 	logger.Log.Debugln("removing test.db")
// 	return os.Remove("./test.db")
// }
