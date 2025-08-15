package repository

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
	"go.uber.org/zap"
)

// keep logger safe in tests
func init() { logger.Log = zap.NewNop().Sugar() }

func TestSaveManga(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewSqlite3Database(dbPath)
	if err != nil {
		t.Fatalf("NewSqlite3Database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ch := model.Chapter{
		Title:      "chapter 10",
		Url:        "https://example.com/berserk/ch10",
		ReleasedAt: time.Now(),
	}
	mg := model.Manga{
		Title:       "Berserk",
		Url:         "https://example.com/berserk",
		LastChapter: &ch,
	}
	chatID := model.ChatID(1111)

	if err := db.MangaRepo.SaveManga(&mg, chatID); err != nil {
		t.Fatalf("SaveManga: %v", err)
	}

	// manga exists
	var mCount int
	if err := db.db.QueryRow(`SELECT COUNT(*) FROM mangas WHERE url = ?`, mg.Url).Scan(&mCount); err != nil {
		t.Fatalf("count mangas: %v", err)
	}
	if mCount != 1 {
		t.Fatalf("want 1 manga, got %d", mCount)
	}

	// last_chapter points to chapter URL
	var lc string
	if err := db.db.QueryRow(`SELECT last_chapter FROM mangas WHERE url = ?`, mg.Url).Scan(&lc); err != nil {
		t.Fatalf("select last_chapter: %v", err)
	}
	if lc != ch.Url {
		t.Fatalf("last_chapter mismatch: got %q want %q", lc, ch.Url)
	}

	// chapter exists
	var cCount int
	if err := db.db.QueryRow(`SELECT COUNT(*) FROM chapters WHERE url = ?`, ch.Url).Scan(&cCount); err != nil {
		t.Fatalf("count chapters: %v", err)
	}
	if cCount != 1 {
		t.Fatalf("want 1 chapter, got %d", cCount)
	}
}