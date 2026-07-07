package executor

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"drug_trials_tool/claw/internal/client"
	"drug_trials_tool/claw/internal/crawler"
	"drug_trials_tool/pkg/store"
)

type Config struct {
	Keyword    string
	Pages      int
	ZoneID     int
	CookieFile string
	DataDir    string
	DBPath     string
	Store      *store.Store
}

type Result struct {
	Found    int
	NewItems int
	Status   string
	Error    string
}

func Execute(cfg Config) *Result {
	r := &Result{}

	st := cfg.Store
	if st == nil {
		var err error
		st, err = store.New(cfg.DBPath)
		if err != nil {
			r.Status = "failed"
			r.Error = "store: " + err.Error()
			return r
		}
		defer st.Close()
	}

	cl := client.New()
	if cfg.CookieFile != "" {
		for _, pair := range strings.Split(cfg.CookieFile, ",") {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				cl.SetCookie(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}
	if err := cl.InitSession(); err != nil {
		r.Status = "failed"
		r.Error = "session: " + err.Error()
		return r
	}

	crawlLog, err := st.StartCrawlLog(cfg.Keyword, cfg.Pages, cfg.ZoneID)
	if err != nil {
		r.Status = "failed"
		r.Error = "crawl_log: " + err.Error()
		return r
	}

	cr := crawler.New(cl)
	pageDir := filepath.Join(cfg.DataDir, "pages")
	os.MkdirAll(pageDir, 0755)

	trials, err := cr.SearchAllPages(cfg.Keyword, cfg.Pages)
	if err != nil {
		st.FinishCrawlLog(crawlLog.ID, 0, 0, "failed")
		r.Status = "failed"
		r.Error = "search: " + err.Error()
		return r
	}

	r.Found = len(trials)
	newItems := 0

	for i, t := range trials {
		log.Printf("[%d/%d] %s (%s)", i+1, len(trials), t.RegNo, t.Title)

		result, err := cr.FetchAndParseDetail(t.DetailID, t.RegNo, cfg.Keyword, pageDir)
		if err != nil {
			log.Printf("  Error: %v", err)
			continue
		}

		if cfg.ZoneID > 0 {
			result.Trial.DiseaseZoneID = cfg.ZoneID
		}

		isNew, err := st.SaveTrial(result.Trial)
		if err != nil {
			log.Printf("  Save error: %v", err)
			continue
		}
		if isNew {
			newItems++
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

	r.NewItems = newItems
	r.Status = "completed"
	if newItems == 0 {
		r.Status = "no_new"
	}
	st.FinishCrawlLog(crawlLog.ID, r.Found, r.NewItems, r.Status)

	log.Printf("Crawl done: %d found, %d new", r.Found, r.NewItems)
	return r
}
