package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	cfg := Configuration{
		headless:    true,
		isOptimized: false,
		browserType: Chromium,
	}

	scraper, err := NewWeebCentralScraper(cfg)
	if err != nil {
		log.Fatalln("FATAL ERROR")
		return
	}

	mangas, err := scraper.FindListOfMangas("One piece")
	if err != nil {
		log.Fatalln("FATAL ERROR")
	}

	manga := mangas[0]
	scraper.FindListOfChapters(manga.url, 3)

}
