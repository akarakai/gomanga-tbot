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

	scraper, err := NewWebCentralScraper(cfg)
	if err != nil {
		log.Fatalln("FATAL ERROR")
		return
	}

	mangas, _ := scraper.FindListOfMangas("One piece");

	manga := mangas[0];
	scraper.FindListOfChapters(manga.url, 3);
	


}	

