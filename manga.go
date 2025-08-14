package main

import "time"

// Basic Manga struct with no behaviour
	type Manga struct {
		title string
		url   string
	}

type Chapter struct {
	title      string
	url        string
	mangaUrl   string
	releasedAt time.Time
}
