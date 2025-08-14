package scraper

import (
	"log"
	"testing"
)

// TODO avoid to open a new instance of the scraper each time
// maybe the logic should be done in the scraper itself
var cachedScraper *PlaywrightScraper

func getScraper() *PlaywrightScraper {
	if cachedScraper != nil {
		return cachedScraper
	}

	cfg := Configuration{
		headless:    false,
		isOptimized: false,
		browserType: Chromium,
	}

	s, err := NewWeebCentralScraper(cfg)
	if err != nil {
		log.Fatalf("failed to initialize scraper: %v", err)
	}

	cachedScraper = s
	return s
}

func TestNewWeebCentralScraper(t *testing.T) {
	scraper := getScraper()

	if scraper == nil {
		t.Fatal("Expected scraper instance, got nil")
	}
}

func TestFindListOfMangas(t *testing.T) {
	s := getScraper()

	t.Run("GoodQuery", func(t *testing.T) {
		query := "Naruto"
		mangas, err := s.FindListOfMangas(query)
		if err != nil {
			t.Errorf("Failed to find list of mangas: %v", err)
		}
		if len(mangas) == 0 {
			t.Errorf("No mangas found. Even if Naruto should exist")
		}
	})

	t.Run("EmptyQuery", func(t *testing.T) {
		bad := ""
		mangas, err := s.FindListOfMangas(bad)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if mangas != nil {
			t.Error("Expected nil mangas")
		}
	})
}

func TestFindListOfChapters_Extra(t *testing.T) {
	s := getScraper()
	validMangaUrl := "https://weebcentral.com/series/01J76XYFXM8RHFVVCN0PJBPAT8/Hikaru-ga-Shinda-Natsu"

	t.Run("ZeroChaptersRequested", func(t *testing.T) {
		chapters, err := s.FindListOfChapters(validMangaUrl, 0)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(chapters) != 0 {
			t.Errorf("Expected 0 chapters, got %d", len(chapters))
		}
	})

	t.Run("RequestMoreChaptersThanAvailable", func(t *testing.T) {
		nChaps := 1000
		chapters, err := s.FindListOfChapters(validMangaUrl, nChaps)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(chapters) > nChaps {
			t.Errorf("Returned more chapters (%d) than requested (%d)", len(chapters), nChaps)
		}
	})

	t.Run("EmptyMangaURL", func(t *testing.T) {
		_, err := s.FindListOfChapters("", 5)
		if err == nil {
			t.Errorf("Expected error for empty URL, got nil")
		}
	})

	t.Run("NilPageInitialization", func(t *testing.T) {
		s.page = nil
		chapters, err := s.FindListOfChapters(validMangaUrl, 2)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(chapters) == 0 {
			t.Errorf("Expected some chapters, got 0")
		}
	})
}

func TestFindImgUrlsOfChapter(t *testing.T) {
	s := getScraper()
	t.Run("GoodQuery", func(t *testing.T) {
		chapterURL := "https://weebcentral.com/chapters/01J76XYYGMWHPGZ0EW6T7BAJKA"
		url, err := s.FindImgUrlsOfChapter(chapterURL)
		if err != nil {
			t.Errorf("Failed to find urls of the chapter: %v", err)
		}
		if len(url) == 0 {
			t.Errorf("No urls found")
		}
	})

	t.Run("EmptyChapterURL", func(t *testing.T) {
		chapterURL := ""
		_, err := s.FindImgUrlsOfChapter(chapterURL)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("WrongBaseUrl", func(t *testing.T) {
		chapterURL := "https://mangadex.com/chapters/01J76XYYGMWHPGZ0EW6T7BAJKA"
		_, err := s.FindImgUrlsOfChapter(chapterURL)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
