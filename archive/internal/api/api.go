package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"drug_trials_tool/internal/crawler"
	"drug_trials_tool/internal/models"
	"drug_trials_tool/internal/store"
)

// Server wraps the REST API server.
type Server struct {
	store    *store.Store
	crawler  *crawler.Crawler
	dataDir  string
	crawling bool
}

// NewServer creates a new API server.
func NewServer(s *store.Store, cr *crawler.Crawler, dataDir string) *Server {
	return &Server{store: s, crawler: cr, dataDir: dataDir}
}

// jsonResp writes a JSON response.
func jsonResp(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: status < 400, Data: data})
}

// jsonErr writes a JSON error response.
func jsonErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: false, Error: msg})
}

// RegisterRoutes sets up all API routes and returns the handler.
func (srv *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// CORS preflight
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
	})

	mux.HandleFunc("/api/provinces", srv.handleProvinces)
	mux.HandleFunc("/api/provinces/", srv.handleProvinceInstitutions)
	mux.HandleFunc("/api/institutions/", srv.handleInstitutionInvestigators)
	mux.HandleFunc("/api/investigators/", srv.handleInvestigatorTrials)
	mux.HandleFunc("/api/trials/", srv.handleTrialDetail)
	mux.HandleFunc("/api/search", srv.handleSearch)
	mux.HandleFunc("/api/crawl/status", srv.handleCrawlStatus)
	mux.HandleFunc("/api/crawl/start", srv.handleCrawlStart)

	return mux
}

// ── Handlers ─────────────────────────────────────────────

func (srv *Server) handleProvinces(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	provinces, err := srv.store.ListProvinces()
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, provinces)
}

func (srv *Server) handleProvinceInstitutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/provinces/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		jsonErr(w, 400, "missing province id")
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		jsonErr(w, 400, "invalid province id")
		return
	}
	institutions, err := srv.store.ListInstitutionsByProvince(id)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, institutions)
}

func (srv *Server) handleInstitutionInvestigators(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/institutions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		jsonErr(w, 400, "missing institution id")
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		jsonErr(w, 400, "invalid institution id")
		return
	}
	investigators, err := srv.store.ListInvestigatorsByInstitution(id)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, investigators)
}

func (srv *Server) handleInvestigatorTrials(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/investigators/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		jsonErr(w, 400, "missing investigator id")
		return
	}
	// parts[0] is investigator ID, but we need name + instID from store
// parts[0] is the investigator ID

	// We need to look up the investigator's name and institution_id
	// Since the store doesn't have a direct function for this, we query via
	// institution's investigators and match by ID
	// For simplicity, let's examine the trials by investigator name + instID from a query param
	// Better approach: use search with investigator parameter
	name := r.URL.Query().Get("name")
	instIDStr := r.URL.Query().Get("institution_id")
	if name == "" || instIDStr == "" {
		jsonErr(w, 400, "need name and institution_id query params")
		return
	}
	instID, _ := strconv.Atoi(instIDStr)
	trials, err := srv.store.ListTrialsByInvestigator(name, instID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, trials)
}

func (srv *Server) handleTrialDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/trials/")
	if id == "" {
		jsonErr(w, 400, "missing trial id")
		return
	}
	trial, err := srv.store.GetTrial(id)
	if err != nil {
		jsonErr(w, 404, "trial not found")
		return
	}
	jsonResp(w, 200, trial)
}

func (srv *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query := models.SearchQuery{
		Keyword:      q.Get("q"),
		Province:     q.Get("province"),
		Institution:  q.Get("institution"),
		Investigator: q.Get("investigator"),
		RegNo:        q.Get("reg_no"),
		Applicant:    q.Get("applicant"),
		Page:         page,
		PageSize:     pageSize,
	}
	result, err := srv.store.SearchTrials(query)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, result)
}

func (srv *Server) handleCrawlStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	sessions, err := srv.store.ListCrawlSessions(20)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, map[string]interface{}{
		"sessions":   sessions,
		"crawling":   srv.crawling,
	})
}

func (srv *Server) handleCrawlStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonErr(w, 405, "method not allowed")
		return
	}
	if srv.crawling {
		jsonErr(w, 409, "crawl already in progress")
		return
	}

	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		jsonErr(w, 400, "missing keyword")
		return
	}
	pagesStr := r.URL.Query().Get("pages")
	pages, _ := strconv.Atoi(pagesStr)
	if pages < 1 {
		pages = 1
	}

	// Start crawl in background
	srv.crawling = true
	go srv.runCrawl(keyword, pages)

	jsonResp(w, 202, map[string]string{
		"message": "crawl started",
		"keyword": keyword,
	})
}

func (srv *Server) runCrawl(keyword string, pages int) {
	defer func() { srv.crawling = false }()

	session := &models.CrawlSession{
		Keyword:      keyword,
		PagesCrawled: 0,
		TrialsFound:  0,
		TrialsNew:    0,
		StartTime:    time.Now(),
		Status:       "running",
	}
	srv.store.StartCrawlSession(keyword)

	log.Printf("Crawl started: keyword=%s pages=%d", keyword, pages)

	// Search all pages
	trials, err := srv.crawler.SearchAllPages(keyword, pages)
	if err != nil {
		log.Printf("Crawl search error: %v", err)
		session.Status = "failed"
		srv.store.UpdateCrawlSession(session)
		return
	}
	session.TrialsFound = len(trials)
	session.PagesCrawled = pages

	// Fetch detail page for each trial
	pageDir := filepath.Join(srv.dataDir, "pages")
	for i, t := range trials {
		log.Printf("  [%d/%d] Fetching detail: %s (%s)", i+1, len(trials), t.RegNo, t.Title)
		result, err := srv.crawler.FetchAndParseDetail(t.DetailID, t.RegNo, keyword, pageDir)
		if err != nil {
			log.Printf("  Error: %v", err)
			continue
		}
		result.Trial.Keyword = keyword

		// Save trial (dedup via hash)
		isNew, err := srv.store.SaveTrial(result.Trial)
		if err != nil {
			log.Printf("  Save error: %v", err)
			continue
		}
		if isNew {
			session.TrialsNew++
		}

		// Process institutions
		for _, ir := range result.Institutions {
			provinceID, _ := srv.store.EnsureProvince(ir.Province)
			instID, _ := srv.store.EnsureInstitution(ir.Institution, ir.City, provinceID)
			if instID > 0 {
				srv.store.LinkTrialInstitution(t.DetailID, instID, ir.Investigator)
				// Also ensure the investigator exists
				if ir.Investigator != "" {
					srv.store.EnsureInvestigator(ir.Investigator, instID, "", "", "", "", "", "")
				}
			}
		}

		// Process main investigator
		if result.Investigator != nil && result.Investigator.Name != "" {
			instName := result.Investigator.InstitutionName
			instID := 0
			if instName != "" {
				// Try to find existing institution by name
				// Use empty province, update later if we have it
				instID, _ = srv.store.EnsureInstitution(instName, "", 0)
			}
			if instID > 0 {
				srv.store.EnsureInvestigator(
					result.Investigator.Name, instID,
					result.Investigator.Degree, result.Investigator.Title,
					result.Investigator.Phone, result.Investigator.Email,
					result.Investigator.Address, result.Investigator.ZipCode)
			}
		}
	}

	session.EndTime = time.Now()
	session.Status = "completed"
	srv.store.UpdateCrawlSession(session)
	log.Printf("Crawl completed: %d trials, %d new", session.TrialsFound, session.TrialsNew)
}

// Serve starts the HTTP server.
func (srv *Server) Serve(addr string) error {
	log.Printf("API server starting on %s", addr)
	return http.ListenAndServe(addr, srv.Handler())
}

