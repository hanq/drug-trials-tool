package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"drug_trials_tool/internal/client"
	"drug_trials_tool/internal/crawler"
	"drug_trials_tool/internal/downloader"
	"drug_trials_tool/internal/store"
)

func main() {
	var (
		keyword     string
		outputDir   string
		cookieFile  string
		ids         string
		maxTrials   int
		pages       int
		saveHTML    string
		crawlKeyword string
		crawlPages   int
		exportJSON  string
		dbPath      string
		dataDir     string
	)

	flag.StringVar(&keyword, "keyword", "", "Disease keyword to search")
	flag.StringVar(&outputDir, "output", "./downloads", "Output directory for DOC files")
	flag.StringVar(&cookieFile, "cookies", "", "Comma-separated cookies")
	flag.StringVar(&ids, "ids", "", "Comma-separated trial IDs to download")
	flag.StringVar(&saveHTML, "save-html", "", "Save search HTML to file")
	flag.IntVar(&maxTrials, "max", 10, "Max trials to download (0=all)")
	flag.IntVar(&pages, "pages", 1, "Search result pages to fetch")
	flag.StringVar(&crawlKeyword, "crawl", "", "Crawl disease keyword and store in DB")
	flag.IntVar(&crawlPages, "crawl-pages", 1, "Pages to crawl")
	flag.StringVar(&exportJSON, "export", "", "Export cleaned data as JSON to path")
	flag.StringVar(&dbPath, "db", "./data/db/trials.db", "SQLite database path")
	flag.StringVar(&dataDir, "data-dir", "./data", "Data directory for pages/ and db/")

	flag.Parse()

	if exportJSON != "" {
		fmt.Println("=== Data Cleaning & Export ===")
		st, err := store.New(dbPath)
		if err != nil {
			log.Fatalf("DB open: %v", err)
		}
		defer st.Close()

		fmt.Println("  Cleaning trial data...")
		trials, err := st.ListAllTrials()
		if err != nil {
			log.Fatalf("List: %v", err)
		}
		updated := 0
		for _, t := range trials {
			cleaned := false
			if s := strings.TrimSpace(t.Title); s != t.Title { t.Title = s; cleaned = true }
			if s := strings.TrimSpace(t.DrugName); s != t.DrugName { t.DrugName = s; cleaned = true }
			if s := strings.TrimSpace(t.Status); s != t.Status { t.Status = s; cleaned = true }
			if s := strings.TrimSpace(t.Indication); s != t.Indication { t.Indication = s; cleaned = true }
			if cleaned { st.UpdateTrial(t); updated++ }
		}
		fmt.Printf("  Cleaned %d / %d trials\n", updated, len(trials))

		fmt.Println("  Building export data...")
		exportData := map[string]interface{}{"generated_at": time.Now().Format("2006-01-02T15:04:05")}

		cnt, _ := st.CountProvinces()
		fmt.Printf("  DB provinces: %d\\n", cnt)
		provinces, err := st.ListProvinces2()
		if err != nil {
			fmt.Printf("  ListProvinces Error: %v\\n", err)
		}
		provList := []map[string]interface{}{}
		type provEntry struct{ id int; name, code string; count int }
		provCountMap := map[string]int{}   // normalized name -> accumulated trial count
		provBestID := map[string]int{}     // normalized name -> id with highest trial_count
		provBestCnt := map[string]int{}    // highest trial_count seen per normalized name
		for _, p := range provinces {
			if !isChineseProvince(p.Name) { continue }
			normName := normalizeProvinceName(p.Name)
			if normName == "" { continue }
			provCountMap[normName] += p.TrialCount
			if p.TrialCount > provBestCnt[normName] {
				provBestCnt[normName] = p.TrialCount
				provBestID[normName] = p.ID
			}
		}
		var provEntries []provEntry
		for _, p := range provinces {
			if !isChineseProvince(p.Name) { continue }
			normName := normalizeProvinceName(p.Name)
			if provBestID[normName] == p.ID {
				provEntries = append(provEntries, provEntry{p.ID, normName, p.Code, provCountMap[normName]})
				delete(provBestID, normName)
			}
		}
		// Build province list from provEntries
		provList = []map[string]interface{}{}
		for _, pe := range provEntries {
			provList = append(provList, map[string]interface{}{
				"id": pe.id, "name": pe.name, "code": pe.code, "trial_count": pe.count,
			})
		}
		exportData["provinces"] = provList

		// Build institution/investigator lists using canonical province entries
		instList := []map[string]interface{}{}
		invList := []map[string]interface{}{}
		for _, pe := range provEntries {
			insts, _ := st.ListInstitutionsByProvince(pe.id)
			for _, inst := range insts {
				instList = append(instList, map[string]interface{}{ "id": inst.ID, "name": inst.Name, "province_id": inst.ProvinceID, "city": inst.City, "trial_count": inst.TrialCount })
				inv, _ := st.ListInvestigatorsByInstitution(inst.ID)
				for _, v := range inv {
					invList = append(invList, map[string]interface{}{ "id": v.ID, "name": v.Name, "degree": v.Degree, "title": v.Title, "phone": v.Phone, "email": v.Email, "institution_id": v.InstitutionID, "trial_count": v.TrialCount })
				}
			}
		}

		trialList := []map[string]interface{}{} 
		for _, t := range trials {
				var dj interface{}
				if t.DetailJSON != nil { json.Unmarshal(t.DetailJSON, &dj) }
			trialList = append(trialList, map[string]interface{}{
				"detail_id": t.DetailID, "reg_no": t.RegNo, "title": t.Title,
				"drug_name": t.DrugName, "indication": t.Indication,
				"status": t.Status, "applicant_name": t.ApplicantName,
				"detail_json": dj,
			})
		}
	exportData["trials"] = trialList
		exportData["institutions"] = instList
		exportData["investigators"] = invList

		links, _ := st.ListAllTrialInstitutions()
		linkList := []map[string]interface{}{} 
		for _, l := range links {
			linkList = append(linkList, map[string]interface{}{
				"trial_id": l.TrialID, "institution_id": l.InstitutionID,
				"investigator_name": l.InvestigatorName,
			})
		}
		exportData["trial_institutions"] = linkList

		jsonData, _ := json.MarshalIndent(exportData, "", "  ")
		os.MkdirAll(filepath.Dir(exportJSON), 0755)
		os.WriteFile(exportJSON, jsonData, 0644)
		// Filter only Chinese
		fmt.Printf("  Exported: %d provinces, %d institutions, %d investigators, %d trials, %d links\n",
			len(provList), len(instList), len(invList), len(trialList), len(linkList))
		return
	}

	if crawlKeyword != "" {
		fmt.Println("=== Clinical Trials Crawler ===")
		fmt.Printf("Crawling: %s (%d pages)\n", crawlKeyword, crawlPages)
		os.MkdirAll(filepath.Join(dataDir, "pages"), 0755)
		os.MkdirAll(filepath.Dir(dbPath), 0755)

		cl := client.New()
		if err := cl.InitSession(); err != nil {
			log.Fatalf("Session: %v", err)
		}
		if cookieFile != "" {
			for _, pair := range strings.Split(cookieFile, ",") {
				parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
				if len(parts) == 2 { cl.SetCookie(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])) }
			}
		}

		st, err := store.New(dbPath)
		if err != nil { log.Fatalf("DB: %v", err) }
		defer st.Close()

		cr := crawler.New(cl)
		trials, err := cr.SearchAllPages(crawlKeyword, crawlPages)
		if err != nil { log.Fatalf("Search: %v", err) }

		pageDir := filepath.Join(dataDir, "pages")
		for i, t := range trials {
			fmt.Printf("[%d/%d] %s\n", i+1, len(trials), t.Title)
			result, err := cr.FetchAndParseDetail(t.DetailID, t.RegNo, crawlKeyword, pageDir)
			if err != nil { fmt.Printf("  Error: %v\n", err); continue }

			isNew, err := st.SaveTrial(result.Trial)
			if err != nil { fmt.Printf("  Save error: %v\n", err); continue }
			if isNew { fmt.Printf("  NEW\n") } else { fmt.Printf("  (unchanged)\n") }

			for _, ir := range result.Institutions {
				if pid, _ := st.EnsureProvince(ir.Province); pid > 0 {
					if instID, _ := st.EnsureInstitution(ir.Institution, ir.City, pid); instID > 0 {
						st.LinkTrialInstitution(t.DetailID, instID, ir.Investigator)
						if ir.Investigator != "" { st.EnsureInvestigator(ir.Investigator, instID, "", "", "", "", "", "") }
					}
				}
			}
			if result.Investigator != nil && result.Investigator.Name != "" {
				instName := result.Investigator.InstitutionName
				if instName != "" {
					if instID, _ := st.EnsureInstitution(instName, "", 0); instID > 0 {
						st.EnsureInvestigator(result.Investigator.Name, instID, result.Investigator.Degree, result.Investigator.Title, result.Investigator.Phone, result.Investigator.Email, result.Investigator.Address, result.Investigator.ZipCode)
					}
				}
			}
		}
		fmt.Printf("\nDone: %d trials, DB: %s\n", len(trials), dbPath)
		return
	}

	if keyword == "" && ids == "" {
		fmt.Println("Usage:")
		fmt.Println("  Crawl:   drug_trials_tool -crawl <keyword> [-crawl-pages N] -db <path>")
		fmt.Println("  Export:  drug_trials_tool -export <path> -db <path>")
		fmt.Println("  Download: drug_trials_tool -keyword <k> -output <dir> -pages N")
		os.Exit(0)
	}

	// Download mode
	fmt.Println("=== Clinical Trials Downloader ===")
	cl := client.New()
	if cookieFile != "" {
		for _, pair := range strings.Split(cookieFile, ",") {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 { cl.SetCookie(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])) }
		}
	}
	if err := cl.InitSession(); err != nil { log.Fatalf("Session: %v", err) }

	cr := crawler.New(cl)
	dl := downloader.New(cl, outputDir)

	var trials []crawler.TrialInfo
	var err error
	if pages <= 1 { trials, err = cr.SearchByDisease(keyword) } else { trials, err = cr.SearchAllPages(keyword, pages) }
	if err != nil { log.Fatalf("Search: %v", err) }
	fmt.Printf("Found %d trials\n", len(trials))

	if maxTrials > 0 && maxTrials < len(trials) { trials = trials[:maxTrials] }
	for _, t := range trials {
		result, err := dl.DownloadByID(t.DetailID, t.Title)
		if err != nil { fmt.Printf("  Error: %v\n", err) } else if result.Success { fmt.Printf("  OK: %s\n", result.FilePath) } else { fmt.Printf("  Failed: %s\n", result.Error) }
	}
	fmt.Printf("\nDone: %d downloaded\n", len(trials))
}

func normalizeProvinceName(name string) string {
	normalized := map[string]string{
		"上海": "上海市", "北京": "北京市", "天津": "天津市", "重庆": "重庆市",
		"安徽": "安徽省", "福建": "福建省", "甘肃": "甘肃省", "广东": "广东省",
		"贵州": "贵州省", "海南": "海南省", "河北": "河北省", "河南": "河南省",
		"黑龙江": "黑龙江省", "湖北": "湖北省", "湖南": "湖南省", "吉林": "吉林省",
		"江苏": "江苏省", "江西": "江西省", "辽宁": "辽宁省", "宁夏": "宁夏回族自治区",
		"青海": "青海省", "山东": "山东省", "山西": "山西省", "陕西": "陕西省",
		"四川": "四川省", "新疆": "新疆维吾尔自治区", "西藏": "西藏自治区",
		"云南": "云南省", "浙江": "浙江省", "广西": "广西壮族自治区",
		"香港": "香港特别行政区", "澳门": "澳门特别行政区",
		"内蒙古": "内蒙古自治区",
		"广州省": "广东省", "宁夏省": "宁夏回族自治区", "广西省": "广西壮族自治区",
		"新疆省": "新疆维吾尔自治区", "西藏省": "西藏自治区",
		"内蒙": "内蒙古自治区", "柏市": "", "柏林": "",
	}
	if n, ok := normalized[name]; ok { return n }
	if strings.HasSuffix(name, "省") || strings.HasSuffix(name, "市") ||
		strings.HasSuffix(name, "区") { return name }
	return name
}

func isChineseProvince(name string) bool {
	if name == "" { return false }
	chinese := []string{"北京","上海","天津","重庆","安徽","福建","甘肃","广东","贵州","海南","河北","河南","黑龙江","湖北","湖南","吉林","江苏","江西","辽宁","内蒙古","宁夏","青海","山东","山西","陕西","四川","新疆","西藏","云南","浙江","广西","香港","澳门"}
	for _, c := range chinese {
		if strings.Contains(name, c) { return true }
	}
	excluded := map[string]bool{"柏市": true, "柏林": true}
	if excluded[name] { return false }
	return strings.HasSuffix(name, "省") || strings.HasSuffix(name, "市")
}
