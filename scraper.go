package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/playwright-community/playwright-go"
)

const WEEB_CENTRAL_BASE_URL = "https://weebcentral.com"
const WINDOW_HEIGHT = 400
const WINDOW_WIDTH = 400

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

// NewWebCentralScraper creates a new instance of WeebCentralScraper and returns its pointer.
// this function will also open a new page from a context
func NewWebCentralScraper(cfg Configuration) (*weebCentralScraper, error) {
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

// FindListOfMangas scrapes the list of mangas
//
// Returns an empty slice if no manga is found or if it encounters an error
func (s *weebCentralScraper) FindListOfMangas(query string) ([]Manga, error) {
	const XPATH_MANGAS_CONTAINER = "/html/body/header/section[1]/div[2]/section/div[2]"
	const ID_SEARCH_BOX = "#quick-search-input"

	page, err := s.context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("error creating a page")
	}

	s.page = page

	log.Printf("Going to %s\n", WEEB_CENTRAL_BASE_URL)
	r, err := s.page.Goto(WEEB_CENTRAL_BASE_URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, fmt.Errorf("problem navigating to %s. Status: %s", WEEB_CENTRAL_BASE_URL, r.StatusText())
	}
	if !r.Ok() {
		return nil, fmt.Errorf("problem navigating to %s. Status: %s", WEEB_CENTRAL_BASE_URL, r.StatusText())
	}

	log.Println("Looking for the search bar...")
	// insert the text in the searchbar
	searchBar := s.page.Locator(ID_SEARCH_BOX)
	if err := searchBar.Fill(query); err != nil {
		return nil, err
	}
	container := s.page.Locator(fmt.Sprintf("xpath=%s", XPATH_MANGAS_CONTAINER))
	links := container.Locator("a")
	links.WaitFor()

	locators, err := links.All()
	if err != nil {
		log.Fatalf("Failed to get locators: %v", err)
		return nil, err;
	}

	mangas := make([]Manga, 0, 10);
	for _, aLoc := range locators {
		// for each a tag, find the url and the name
		title, _ := aLoc.InnerText();
		href, _ := aLoc.GetAttribute("href");
		
		mangas = append(mangas, Manga{
			title: title,
			url: href,
		})
	}

	log.Printf("Query \"%s\" gave %d results", query, len(mangas))
	return mangas, nil
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

func (s *weebCentralScraper) Close() error {
	slog.Info("Closing the browser...")
	return s.browser.Close()
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
			Width:  WINDOW_HEIGHT,
			Height: WINDOW_WIDTH,
		},
	}

}
