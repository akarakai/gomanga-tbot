package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	cfg := Configuration{
		headless:    false,
		isOptimized: false,
		browserType: Chromium,
	}

	scraper, err := NewWebCentralScraper(cfg)
	if err != nil {
		log.Fatalln("FATAL ERROR")
		return
	}

	scraper.FindListOfMangas("One piece")
}
