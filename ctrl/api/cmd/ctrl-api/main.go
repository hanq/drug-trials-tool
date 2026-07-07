package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"drug_trials_tool/pkg/store"
)

func main() {
	var (
		port   int
		dbPath string
	)
	flag.IntVar(&port, "port", 8081, "HTTP server port")
	flag.StringVar(&dbPath, "db", "../data/trials.db", "SQLite database path")
	flag.Parse()

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Store: %v", err)
	}
	defer st.Close()

	mux := http.NewServeMux()
	srv := &adminServer{store: st}

	mux.HandleFunc("/api/ctrl/health", srv.health)

	// Auth
	mux.HandleFunc("/api/ctrl/login", srv.handleLogin)

	// Dashboard
	mux.HandleFunc("/api/ctrl/dashboard", srv.handleDashboard)

	// Trials
	mux.HandleFunc("/api/ctrl/trials", srv.handleTrials)
	mux.HandleFunc("/api/ctrl/trials/", srv.handleTrialAction)

	// Disease Zones
	mux.HandleFunc("/api/ctrl/disease-zones", srv.handleDiseaseZones)
	mux.HandleFunc("/api/ctrl/disease-zones/", srv.handleDiseaseZoneDelete)

	// Crawl
	mux.HandleFunc("/api/ctrl/crawl/start", srv.handleCrawlStart)
	mux.HandleFunc("/api/ctrl/crawl/status", srv.handleCrawlStatus)
	mux.HandleFunc("/api/ctrl/crawl/logs", srv.handleCrawlLogs)

	// Announcements
	mux.HandleFunc("/api/ctrl/announcements", srv.handleAnnouncements)
	mux.HandleFunc("/api/ctrl/announcements/", srv.handleAnnouncementByID)

	h := corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Ctrl API on %s (db: %s)", addr, dbPath)
	log.Fatal(http.ListenAndServe(addr, h))
}

type adminServer struct {
	store *store.Store
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" { w.WriteHeader(204); return }
		next.ServeHTTP(w, r)
	})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
}

func jsonErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": msg})
}

func (s *adminServer) health(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}

// Auth

func (s *adminServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { jsonErr(w, 405, "method not allowed"); return }
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json"); return
	}
	admin, err := s.store.GetAdminByUsername(body.Username)
	if err != nil {
		jsonErr(w, 401, "invalid credentials"); return
	}
	h := sha256.Sum256([]byte(body.Password))
	hash := fmt.Sprintf("%x", h)
	if admin.PasswordHash != hash {
		jsonErr(w, 401, "invalid credentials"); return
	}
	jsonOK(w, map[string]interface{}{"id": admin.ID, "username": admin.Username})
}

// Dashboard

func (s *adminServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	stats, err := s.store.DashboardStats()
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, stats)
}

// Trials

func (s *adminServer) handleTrials(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		zoneID, _ := strconv.Atoi(r.URL.Query().Get("zone_id"))
		keyword := r.URL.Query().Get("keyword")
		trials, err := s.store.ListUnpublishedTrials(zoneID, keyword)
		if err != nil { jsonErr(w, 500, err.Error()); return }
		jsonOK(w, trials)
		return
	}
	if r.Method == "POST" {
		// Batch publish
		var body struct {
			ZoneID  int    `json:"zone_id"`
			Keyword string `json:"keyword"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		n, err := s.store.BatchPublish(body.ZoneID, body.Keyword)
		if err != nil { jsonErr(w, 500, err.Error()); return }
		jsonOK(w, map[string]int{"published": n})
		return
	}
	jsonErr(w, 405, "method not allowed")
}

func (s *adminServer) handleTrialAction(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/ctrl/trials/")
	if id == "" { jsonErr(w, 400, "missing trial id"); return }
	if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/publish") {
		trialID := strings.TrimSuffix(id, "/publish")
		adminID, _ := strconv.Atoi(r.URL.Query().Get("admin_id"))
		if adminID == 0 { adminID = 1 }
		if err := s.store.PublishTrial(trialID, adminID); err != nil {
			jsonErr(w, 500, err.Error()); return
		}
		jsonOK(w, map[string]string{"status": "published"})
		return
	}
	jsonErr(w, 405, "method not allowed")
}

// Disease Zones

func (s *adminServer) handleDiseaseZones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		zones, err := s.store.ListDiseaseZones()
		if err != nil { jsonErr(w, 500, err.Error()); return }
		jsonOK(w, zones)
	case "POST":
		var body struct {
			Name        string `json:"name"`
			Keyword     string `json:"keyword"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonErr(w, 400, "invalid json"); return
		}
		zone, err := s.store.CreateDiseaseZone(body.Name, body.Keyword, body.Description)
		if err != nil { jsonErr(w, 500, err.Error()); return }
		jsonOK(w, zone)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

func (s *adminServer) handleDiseaseZoneDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" { jsonErr(w, 405, "method not allowed"); return }
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/ctrl/disease-zones/"))
	if err != nil { jsonErr(w, 400, "invalid id"); return }
	if err := s.store.DeleteDiseaseZone(id); err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, map[string]string{"status": "deleted"})
}

// Crawl

func (s *adminServer) handleCrawlStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { jsonErr(w, 405, "method not allowed"); return }
	var body struct {
		Keyword string `json:"keyword"`
		Pages   int    `json:"pages"`
		ZoneID  int    `json:"zone_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Keyword == "" { jsonErr(w, 400, "missing keyword"); return }
	if body.Pages < 1 { body.Pages = 1 }

	log.Printf("Crawl triggered: keyword=%s pages=%d zone=%d", body.Keyword, body.Pages, body.ZoneID)
	jsonOK(w, map[string]interface{}{
		"message": "crawl queued (run claw CLI separately)",
		"keyword": body.Keyword,
		"pages":   body.Pages,
	})
}

func (s *adminServer) handleCrawlStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{"crawling": false, "message": "use POST /api/ctrl/crawl/start to trigger"})
}

func (s *adminServer) handleCrawlLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	logs, err := s.store.ListCrawlLogs(20)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, logs)
}

// Announcements

func (s *adminServer) handleAnnouncements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		anns, err := s.store.ListAnnouncements(false)
		if err != nil { jsonErr(w, 500, err.Error()); return }
		jsonOK(w, anns)
	case "POST":
		var body struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonErr(w, 400, "invalid json"); return
		}
		ann, err := s.store.CreateAnnouncement(body.Title, body.Content, 1)
		if err != nil { jsonErr(w, 500, err.Error()); return }
		jsonOK(w, ann)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

func (s *adminServer) handleAnnouncementByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/ctrl/announcements/"))
	if err != nil { jsonErr(w, 400, "invalid id"); return }
	switch r.Method {
	case "PUT":
		var body struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if err := s.store.UpdateAnnouncement(id, body.Title, body.Content); err != nil {
			jsonErr(w, 500, err.Error()); return
		}
		jsonOK(w, map[string]string{"status": "updated"})
	case "DELETE":
		if err := s.store.DeleteAnnouncement(id); err != nil {
			jsonErr(w, 500, err.Error()); return
		}
		jsonOK(w, map[string]string{"status": "deleted"})
	default:
		jsonErr(w, 405, "method not allowed")
	}
}
