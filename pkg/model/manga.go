package model

import (
	"sort"
	"time"
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

// SortMangaByRecentChapter sorts manga based on the most recent chapter's ReleasedAt date, closest to the present time.
func SortMangaByRecentChapter(mangaList []Manga) {
	sort.SliceStable(mangaList, func(i, j int) bool {
		return mangaList[i].LastChapter.ReleasedAt.After(mangaList[j].LastChapter.ReleasedAt)
	})
}
