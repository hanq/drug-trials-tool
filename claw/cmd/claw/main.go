package main

import (
	"flag"
	"fmt"
	"log"

	"drug_trials_tool/claw/executor"
)

func main() {
	var (
		keyword    string
		dbPath     string
		pages      int
		zoneID     int
		cookieFile string
		dataDir    string
	)
	flag.StringVar(&keyword, "keyword", "", "Disease keyword to crawl")
	flag.StringVar(&dbPath, "db", "../../data/trials.db", "SQLite database path")
	flag.IntVar(&pages, "pages", 1, "Search result pages to fetch")
	flag.IntVar(&zoneID, "zone", 0, "Disease zone ID to assign")
	flag.StringVar(&cookieFile, "cookies", "", "Comma-separated cookies")
	flag.StringVar(&dataDir, "data-dir", "../../data", "Data directory for pages/")
	flag.Parse()

	if keyword == "" {
		fmt.Println("Usage: claw -keyword <disease> [-pages N] [-db <path>] [-zone N]")
		fmt.Println("Example: claw -keyword \"胰腺\" -pages 3 -zone 1")
		return
	}

	log.Printf("Starting crawl: keyword=%s pages=%d zone=%d", keyword, pages, zoneID)
	result := executor.Execute(executor.Config{
		Keyword:    keyword,
		Pages:      pages,
		ZoneID:     zoneID,
		CookieFile: cookieFile,
		DataDir:    dataDir,
		DBPath:     dbPath,
	})

	if result.Status == "failed" {
		log.Fatalf("Crawl failed: %s", result.Error)
	}

	fmt.Printf("\n=== Crawl Complete ===\n")
	fmt.Printf("  Found:   %d trials\n", result.Found)
	fmt.Printf("  New:     %d trials\n", result.NewItems)
	fmt.Printf("  Status:  %s\n", result.Status)
}
