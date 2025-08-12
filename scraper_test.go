package main

import (
	"log"
	"testing"
)

// TODO controllare nello scraper se link inizia con weebcentral
func TestFindListOfChapters(t *testing.T) {
	s := getScraper()
	mangaUrl := "https://weebcentral.com/series/01J76XYFXM8RHFVVCN0PJBPAT8/Hikaru-ga-Shinda-Natsu"
	nChaps := 3

	t.Run("ChaptersFound", func(t *testing.T) {
		chaps, err := s.FindListOfChapters(mangaUrl, nChaps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(chaps) != nChaps {
			t.Errorf("Expected %d chapters, got %d", nChaps, len(chaps))
		}

		if len(chaps) > 0 && chaps[0].mangaUrl != mangaUrl {
			t.Errorf("Expected mangaUrl %s, got %s", mangaUrl, chaps[0].mangaUrl)
		}
	})

	t.Run("InvalidMangaUrl", func(t *testing.T) {
		mangaUrl := mangaUrl + "hehe"

	})

}

func getScraper() *weebCentralScraper {
	cfg := Configuration{
		headless:    true,
		isOptimized: false,
		browserType: Chromium,
	}

	s, err := NewWebCentralScraper(cfg)
	if err != nil {
		log.Fatalf("failed to initialize scraper: %v", err)
	}
	return s
}
