package main

import (
    "flag"
    "fmt"
    "log"
)

func main() {
    var (
        keyword    string
        dbPath     string
        pages      int
        zoneID     int
        cookieFile string
    )
    flag.StringVar(&keyword, "keyword", "", "Disease keyword to crawl")
    flag.StringVar(&dbPath, "db", "../data/trials.db", "SQLite database path")
    flag.IntVar(&pages, "pages", 1, "Search result pages to fetch")
    flag.IntVar(&zoneID, "zone", 0, "Disease zone ID to assign")
    flag.StringVar(&cookieFile, "cookies", "", "Comma-separated cookies")
    flag.Parse()

    if keyword == "" {
        fmt.Println("Usage: claw -keyword <disease> [-pages N] [-db <path>] [-zone N]")
        return
    }

    log.Printf("Claw started: keyword=%s, pages=%d, zone=%d", keyword, pages, zoneID)
    log.Printf("Claw completed: 0 trials found (TODO: implement crawler)")
}
