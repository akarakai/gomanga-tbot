package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

const WeebCentralBaseURL = "https://weebcentral.com"
const WindowHeight = 400
const WindowWidth = 400

// Browser type as ENUM
type BrowserType int

const (
	Chromium BrowserType = iota
	Firefox
	Webkit
)

type weebCentralScraper struct {
	pw      *playwright.Playwright
	browser playwright.Browser        // interface
	context playwright.BrowserContext // interface
	page    playwright.Page           // interface
	cfg     Configuration
}

type Configuration struct {
	headless bool
	// remove many things to speed up
	isOptimized bool
	// firefox, chromium or webkit are supported
	browserType BrowserType
}

// NewWeebCentralScraper creates a new instance of WeebCentralScraper and returns its pointer.
// this function will also open a new page from a context
func NewWeebCentralScraper(cfg Configuration) (*weebCentralScraper, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	browserOpt := getBrowserOptions(cfg)
	var browser playwright.Browser
	switch cfg.browserType {
	case Chromium:
		browser, err = pw.Chromium.Launch(*browserOpt)
	case Firefox:
		browser, err = pw.Firefox.Launch(*browserOpt)
	case Webkit:
		browser, err = pw.WebKit.Launch(*browserOpt)
	default:
		return nil, errors.New("browser is not supported")
	}
	if err != nil {
		return nil, err
	}

	context, err := browser.NewContext(getContextOptions())
	if err != nil {
		return nil, err
	}
	log.Println("scraper created successfully")
	return &weebCentralScraper{
		pw:      pw,
		browser: browser,
		context: context,
		page:    nil,
		cfg:     cfg,
	}, nil
}

// No need to pass configuration
func NewWeebCentralScraperDefault() (*weebCentralScraper, error) {
	cfg := Configuration{
		headless:    true,
		isOptimized: false,
		browserType: Chromium,
	}

	return NewWeebCentralScraper(cfg)
}

// FindListOfMangas scrapes the list of mangas
//
// Returns an empty slice if no manga is found or if it encounters an error
// the retourned mangas do not contain the last chapter, for that you must use the FindListOfChapters
// with nChaps = 1
func (s *weebCentralScraper) FindListOfMangas(query string) ([]Manga, error) {
	if query == "" {
		log.Println("query is empty")
		return nil, errors.New("query is empty")
	}
	const XPATH_MANGAS_CONTAINER = "/html/body/header/section[1]/div[2]/section/div[2]"
	const ID_SEARCH_BOX = "#quick-search-input"

	if s.page == nil {
		log.Println("Page is empty. Creating a new one.")
		if err := makeNewPage(s); err != nil {
			return nil, err
		}
	}

	log.Printf("Going to %s\n", WeebCentralBaseURL)
	r, err := s.page.Goto(WeebCentralBaseURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, fmt.Errorf("problem navigating to %s", WeebCentralBaseURL)
	}
	if !r.Ok() {
		return nil, fmt.Errorf("problem navigating to %s. Status: %s", WeebCentralBaseURL, r.StatusText())
	}

	log.Println("Looking for the search bar...")
	// insert the text in the searchbar
	searchBar := s.page.Locator(ID_SEARCH_BOX)
	if err := searchBar.Fill(query); err != nil {
		return nil, err
	}
	container := s.page.Locator(fmt.Sprintf("xpath=%s", XPATH_MANGAS_CONTAINER))
	err = container.WaitFor()
	if err != nil {
		return nil, err
	}
	links := container.Locator("a")

	locators, err := links.All()
	if err != nil {
		log.Fatalf("Failed to get locators: %v", err)
		return nil, err
	}

	mangas := make([]Manga, 0, 10)
	for _, aLoc := range locators {
		// for each a tag, find the url and the name
		title, _ := aLoc.InnerText()
		href, _ := aLoc.GetAttribute("href")
		mangas = append(mangas, Manga{
			title:       title,
			url:         href,
			lastChapter: nil,
		})
	}

	log.Printf("Query \"%s\" gave %d results", query, len(mangas))
	return mangas, nil
}

// FindListOfChapters finds the required number of most recent chapters from a manga.
func (s *weebCentralScraper) FindListOfChapters(mangaURL string, nChaps int) ([]Chapter, error) {
	if !strings.HasPrefix(mangaURL, WeebCentralBaseURL) {
		return nil, fmt.Errorf("url %q does not have prefix %q", mangaURL, WeebCentralBaseURL)
	}

	const chapterListSelector = "#chapter-list"

	if s.page == nil {
		log.Println("Page is empty. Creating a new one.")
		if err := makeNewPage(s); err != nil {
			return nil, err
		}
	}

	log.Printf("Navigating to manga URL: %s", mangaURL)
	resp, err := s.page.Goto(mangaURL)
	if err != nil {
		return nil, fmt.Errorf("error navigating to %s: %w", mangaURL, err)
	}
	if !resp.Ok() {
		return nil, fmt.Errorf("received non-OK status %s for URL %s", resp.StatusText(), mangaURL)
	}

	// Check for redirect to 404 page
	if resp.URL() == "https://weebcentral.com/404" {
		return nil, fmt.Errorf("manga with url %s was not found", mangaURL)
	}

	// Locate chapter elements
	chapterDivs, err := s.page.Locator(chapterListSelector).Locator("div").All()
	if err != nil {
		return nil, fmt.Errorf("failed to locate chapters: %w", err)
	}

	// Limit chapters if requested nChaps is less than found chapters
	limit := nChaps
	if limit > len(chapterDivs) {
		limit = len(chapterDivs)
	}
	chapterDivs = chapterDivs[:limit]

	chapters := make([]Chapter, 0, limit)

	for _, chDiv := range chapterDivs {
		title, date, href := extractChapterData(chDiv)
		log.Printf("Extracted chapter: %s", title)
		chapters = append(chapters, Chapter{
			title:      title,
			url:        href,
			mangaUrl:   mangaURL,
			releasedAt: date,
		})
	}

	return chapters, nil
}

func (s *weebCentralScraper) FindImgUrlsOfChapter(chapterURL string) ([]string, error) {
	if !strings.HasPrefix(chapterURL, WeebCentralBaseURL) {
		return nil, fmt.Errorf("url %q does not have prefix %q", chapterURL, WeebCentralBaseURL)
	}
	if chapterURL == "" {
		return nil, fmt.Errorf("chapterURL is empty")
	}
	if s.page == nil {
		log.Println("Page is empty. Creating a new one.")
		if err := makeNewPage(s); err != nil {
			return nil, err
		}
	}

	r, err := s.page.Goto(chapterURL)
	if err != nil {
		return nil, fmt.Errorf("error navigating to %s: %w", chapterURL, err)
	}
	if !r.Ok() {
		return nil, fmt.Errorf("received non-OK status %s for URL %s", r.StatusText(), chapterURL)
	}

	const xpathImgsContainer = "/html/body/main/section[3]"
	container := s.page.Locator(fmt.Sprintf("xpath=%s", xpathImgsContainer))
	err = container.WaitFor()
	if err != nil {
		return nil, fmt.Errorf("container not found: %w", err)
	}
	imgList, err := container.Locator("img").All()
	if err != nil {
		return nil, fmt.Errorf("failed to locate imgs: %w", err)
	}

	imgs := make([]string, 0, len(imgList))
	for _, img := range imgList {
		src, err := img.GetAttribute("src")
		if err != nil {
			imgs = append(imgs, "")
			continue
		}

		log.Println(src)
		imgs = append(imgs, src)
	}

	log.Printf("Found %d images", len(imgs))
	return imgs, nil
}

func (s *weebCentralScraper) CurrentUrl() string {
	if s.page == nil { // this is an interface. TODO check better interface assertion
		return ""
	}
	return s.page.URL()
}

func (s *weebCentralScraper) CurrentPageTitle() (string, error) {
	if s.page == nil { // this is an interface. TODO check better interface assertion
		return "", errors.New("title of the page not found")
	}
	return s.page.Title()
}

func (s *weebCentralScraper) Close() {
	slog.Info("Closing the browser...")
	err := s.browser.Close()
	if err != nil {
		Log.Errorw("failed to close browser", "err", err)
		return
	}
}

func getBrowserOptions(cfg Configuration) *playwright.BrowserTypeLaunchOptions {
	if !cfg.headless {
		return &playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
		}
	}

	// optimized
	return &playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-accelerated-2d-canvas",
			"--disable-gpu",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
			"--disable-features=TranslateUI",
			"--disable-ipc-flooding-protection",
			"--disable-default-apps",
			"--disable-extensions",
			"--disable-sync",
			"--disable-background-networking",
			"--disable-component-update",
			"--no-default-browser-check",
			"--no-first-run",
			"--disable-web-security",
			"--disable-features=VizDisplayCompositor",
		},
	}

}

func getContextOptions() playwright.BrowserNewContextOptions {
	return playwright.BrowserNewContextOptions{
		JavaScriptEnabled: playwright.Bool(true),
		// Images:           playwright.Bool(false), // Disable image loading
		ColorScheme: playwright.ColorSchemeLight,
		Viewport: &playwright.Size{
			Width:  WindowHeight,
			Height: WindowWidth,
		},
	}

}

func extractChapterData(divLoc playwright.Locator) (title string, releasedAt time.Time, href string) {
	aLoc := divLoc.Locator("a")

	href, err := aLoc.GetAttribute("href")
	if err != nil {
		log.Printf("Failed to find href attribute: %v", err)
		href = ""
	}

	// Extract the datetime attribute from <time> inside <a>
	timeLoc := aLoc.Locator("time")
	dateStr, err := timeLoc.GetAttribute("datetime")
	if err != nil {
		log.Printf("Failed to find date attribute: %v", err)
		releasedAt = time.Time{} // zero time
	} else {
		releasedAt, err = time.Parse(time.RFC3339Nano, dateStr)
		if err != nil {
			log.Printf("Failed to parse datetime %q: %v", dateStr, err)
			releasedAt = time.Time{} // zero time
		}
	}

	// Extract title from nested span elements
	titleSpan, err := aLoc.Locator("span").Nth(1).Locator("span").Nth(0).InnerText()
	if err != nil {
		log.Printf("Failed to get title text: %v", err)
		title = ""
	} else {
		title = titleSpan
	}

	return title, releasedAt, href
}

func makeNewPage(s *weebCentralScraper) error {
	page, err := s.context.NewPage()
	if err != nil {
		return err
	}
	s.page = page
	return nil
}
