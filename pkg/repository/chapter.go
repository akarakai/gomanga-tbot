package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type ChapterRepo interface {
	UpdateLastChapter(chapter *model.Chapter, mangaUrl string) error
}

type ChapterRepoSqlite3 struct {
	db *sql.DB
}


func (repo *ChapterRepoSqlite3) UpdateLastChapter(chapter *model.Chapter, mangaUrl string) error {
	_, err := repo.db.Exec(`
		INSERT OR REPLACE INTO chapters (Url, Title, released_at)
		VALUES (?, ?, ?)`,
		chapter.Url, chapter.Title, chapter.ReleasedAt)
	if err != nil {
		return err
	}

	_, err = repo.db.Exec(`
		UPDATE mangas SET last_chapter = ? WHERE Url = ?`,
		chapter.Url, mangaUrl)
	return err
}

func (repo *UserRepoSqlite3) AddMangaToSaved(manga *model.Manga) error {
	_, err := repo.db.Exec(`
		INSERT OR IGNORE INTO mangas (Url, Title) VALUES (?, ?)`,
		manga.Url, manga.Title)
	if err != nil {
		return err
	}

	// Dummy: assuming user_id = 1 for now
	_, err = repo.db.Exec(`
		INSERT OR IGNORE INTO user_mangas (user_id, manga_Url)
		VALUES (?, ?)`, 1, manga.Url)
	return err
}

func (repo *UserRepoSqlite3) RemoveMangaFromSaved(manga *model.Manga) error {
	_, err := repo.db.Exec(`
		DELETE FROM user_mangas WHERE manga_Url = ?`, manga.Url)
	return err
}