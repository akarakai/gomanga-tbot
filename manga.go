package main

import (
	"time"
)

// Basic Manga struct with no behaviour
type Manga struct {
	title       string
	url         string // unique, is ID
	lastChapter *Chapter
}

type Chapter struct {
	title      string
	url        string // unique, is ID
	releasedAt time.Time
}

// for semplicity an user has only a chatID, meaning that if he deletes
// the chat, then he looses the data
type User struct {
	chatID ChatID // int6, unique, is IO
	mangas []Manga
}


