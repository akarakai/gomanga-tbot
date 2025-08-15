package scraper

import (
	"github.com/akarakai/gomanga-tbot/pkg/model"
)

const WeebCentralBaseURL = "https://weebcentral.com"
const WindowHeight = 400
const WindowWidth = 400

type Scraper interface {
	// does not insert last chapter in the manga
	FindListOfMangas(query string) ([]model.Manga, error)
	FindListOfChapters(mangaURL string, nChaps int) ([]model.Chapter, error)
	FindImgUrlsOfChapter(chapterURL string) ([]string, error)
	CurrentUrl() string
	CurrentPageTitle() (string, error)
	Close()
}

type BrowserType int

const (
	Chromium BrowserType = iota
	Firefox
	Webkit
)

type Configuration struct {
	headless    bool
	isOptimized bool
	browserType BrowserType
}
