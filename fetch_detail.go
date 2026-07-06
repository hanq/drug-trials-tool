package main

import (
	"fmt"
	"os"
	"drug_trials_tool/internal/client"
	"drug_trials_tool/internal/crawler"
)

func debugMain() {
	cl := client.New()
	fmt.Println("Initializing session...")
	if err := cl.InitSession(); err != nil {
		fmt.Fprintf(os.Stderr, "Session init failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Session OK")
	cr := crawler.New(cl)
	html, err := cr.GetPageByURL("/clinicaltrials.searchlistdetail.dhtml?id=6728595b4eea44dfa816c8cdfb42c147")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fetch failed: %v\n", err)
		os.Exit(1)
	}
	os.WriteFile("C:\\tmp\\detail_page.html", []byte(html), 0644)
	fmt.Printf("Saved detail page: %d bytes\n", len(html))
}
