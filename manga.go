package main

import "time"

// Basic Manga struct with no behaviour
type Manga struct {
	title 		string
	url   		string
	lastChapter	*Chapter
}

type Chapter struct {
	title      string
	url        string
	releasedAt time.Time
}

// for semplicity an user has only a chatID, meaning that if he deletes
// the chat, then he looses the data
type User struct {
	userID	int64
	chatID 	int64 // maybe to do custom type ChatID and UserID
	mangas  []Manga
}