package model

import (
	"time"
	"sort"
)

type ChatID int64

// Basic Manga struct with no behaviour
type Manga struct {
	Title       string
	Url         string // unique, is ID
	LastChapter *Chapter
}

type Chapter struct {
	Title      string
	Url        string // unique, is ID
	ReleasedAt time.Time
}

// for semplicity an user has only a ChatID, meaning that if he deletes
// the chat, then he looses the data
type User struct {
	ChatID ChatID // int6, unique, is IO
	Mangas []Manga
}

func SortMangasByChapterReleased(mangas []Manga) {
	sort.Slice(mangas, func(i, j int) bool {
		// Both have chapters → compare dates
		if mangas[i].LastChapter != nil && mangas[j].LastChapter != nil {
			return mangas[i].LastChapter.ReleasedAt.After(mangas[j].LastChapter.ReleasedAt)
		}
		// Manga with a chapter goes before one without
		if mangas[i].LastChapter != nil {
			return true
		}
		if mangas[j].LastChapter != nil {
			return false
		}
		// Neither have a chapter → keep current order
		return false
	})
}