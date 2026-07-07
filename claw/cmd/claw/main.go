package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"drug_trials_tool/claw/internal/client"
	"drug_trials_tool/claw/internal/crawler"
	"drug_trials_tool/pkg/store"
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
		fmt.Println("Usage: claw -keyword <disease> [-pages N] [-db <path>] [-zone N] [-cookies k=v,k2=v2]")
		fmt.Println("Example: claw -keyword \"\\u80f0\\u817a\" -pages 3 -zone 1")
		return
	}

	cl := client.New()
	if cookieFile != "" {
		for _, pair := range strings.Split(cookieFile, ",") {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				cl.SetCookie(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}
	if err := cl.InitSession(); err != nil {
		log.Fatalf("Session init: %v", err)
	}

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Store: %v", err)
	}
	defer st.Close()

	crawlLog, err := st.StartCrawlLog(keyword, pages, zoneID)
	if err != nil {
		log.Fatalf("Crawl log: %v", err)
	}

	cr := crawler.New(cl)
	log.Printf("Crawling: keyword=%s pages=%d zone=%d", keyword, pages, zoneID)

	trials, err := cr.SearchAllPages(keyword, pages)
	if err != nil {
		st.FinishCrawlLog(crawlLog.ID, 0, 0, "failed")
		log.Fatalf("Search: %v", err)
	}

	found := len(trials)
	newItems := 0
	pageDir := filepath.Join(dataDir, "pages")
	os.MkdirAll(pageDir, 0755)

	for i, t := range trials {
		log.Printf("[%d/%d] %s (%s)", i+1, len(trials), t.RegNo, t.Title)

		result, err := cr.FetchAndParseDetail(t.DetailID, t.RegNo, keyword, pageDir)
		if err != nil {
			log.Printf("  Error: %v", err)
			continue
		}

		if zoneID > 0 {
			result.Trial.DiseaseZoneID = zoneID
		}

		isNew, err := st.SaveTrial(result.Trial)
		if err != nil {
			log.Printf("  Save error: %v", err)
			continue
		}
		if isNew {
			newItems++
			log.Printf("  NEW")
		} else {
			log.Printf("  (unchanged)")
		}

		for _, ir := range result.Institutions {
			provinceID, _ := st.EnsureProvince(ir.Province)
			instID, _ := st.EnsureInstitution(ir.Institution, ir.City, provinceID)
			if instID > 0 {
				st.LinkTrialInstitution(t.DetailID, instID, ir.Investigator)
				if ir.Investigator != "" {
					st.EnsureInvestigator(ir.Investigator, instID, "", "", "", "", "", "")
				}
			}
		}

		if result.Investigator != nil && result.Investigator.Name != "" {
			instName := result.Investigator.InstitutionName
			instID := 0
			if instName != "" {
				instID, _ = st.EnsureInstitution(instName, "", 0)
			}
			if instID > 0 {
				st.EnsureInvestigator(
					result.Investigator.Name, instID,
					result.Investigator.Degree, result.Investigator.Title,
					result.Investigator.Phone, result.Investigator.Email,
					result.Investigator.Address, result.Investigator.ZipCode)
			}
		}
	}

	status := "completed"
	if newItems == 0 {
		status = "no_new"
	}
	st.FinishCrawlLog(crawlLog.ID, found, newItems, status)
	log.Printf("Done: %d found, %d new (db: %s)", found, newItems, dbPath)
}
