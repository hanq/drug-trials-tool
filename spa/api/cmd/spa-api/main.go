package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"drug_trials_tool/pkg/models"
	"drug_trials_tool/pkg/store"
)

func main() {
	var (
		port   int
		dbPath string
	)
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.StringVar(&dbPath, "db", "../data/trials.db", "SQLite database path")
	flag.Parse()

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Store: %v", err)
	}
	defer st.Close()

	mux := http.NewServeMux()
	srv := &server{store: st}

	mux.HandleFunc("/api/spa/health", srv.health)

	// Public browse
	mux.HandleFunc("/api/spa/provinces", srv.handleProvinces)
	mux.HandleFunc("/api/spa/provinces/", srv.handleProvinceInst)
	mux.HandleFunc("/api/spa/institutions/", srv.handleInstitutionInvest)
	mux.HandleFunc("/api/spa/investigators/", srv.handleInvestigatorTrials)
	mux.HandleFunc("/api/spa/trials/", srv.handleTrialDetail)
	mux.HandleFunc("/api/spa/search", srv.handleSearch)

	// Disease zones
	mux.HandleFunc("/api/spa/disease-zones", srv.handleDiseaseZones)
	mux.HandleFunc("/api/spa/disease-zones/", srv.handleZoneTrials)

	// Announcements
	mux.HandleFunc("/api/spa/announcements", srv.handleAnnouncements)

	// CORS + JSON defaults
	h := corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("SPA API on %s (db: %s)", addr, dbPath)
	log.Fatal(http.ListenAndServe(addr, h))
}

type server struct {
	store *store.Store
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
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

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}

// Provinces
func (s *server) handleProvinces(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	provinces, err := s.store.ListProvinces()
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, provinces)
}

func (s *server) handleProvinceInst(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/spa/provinces/"))
	if err != nil { jsonErr(w, 400, "invalid province id"); return }
	insts, err := s.store.ListInstitutionsByProvince(id)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, insts)
}

func (s *server) handleInstitutionInvest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/spa/institutions/"))
	if err != nil { jsonErr(w, 400, "invalid institution id"); return }
	invs, err := s.store.ListInvestigatorsByInstitution(id)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, invs)
}

func (s *server) handleInvestigatorTrials(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	name := r.URL.Query().Get("name")
	instID, _ := strconv.Atoi(r.URL.Query().Get("institution_id"))
	if name == "" || instID == 0 { jsonErr(w, 400, "need name and institution_id"); return }
	trials, err := s.store.ListTrialsByInvestigator(name, instID)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, trials)
}

func (s *server) handleTrialDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	id := strings.TrimPrefix(r.URL.Path, "/api/spa/trials/")
	if id == "" { jsonErr(w, 400, "missing trial id"); return }
	trial, err := s.store.GetTrial(id)
	if err != nil { jsonErr(w, 404, "trial not found"); return }
	jsonOK(w, trial)
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 { page = 1 }
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }
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
	result, err := s.store.SearchTrials(query)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, result)
}

// Disease Zones

func (s *server) handleDiseaseZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	zones, err := s.store.ListDiseaseZones()
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, zones)
}

func (s *server) handleZoneTrials(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	path := strings.TrimPrefix(r.URL.Path, "/api/spa/disease-zones/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" { jsonErr(w, 400, "missing zone id"); return }
	id, err := strconv.Atoi(parts[0])
	if err != nil { jsonErr(w, 400, "invalid zone id"); return }
	trials, err := s.store.ListTrialsByZone(id)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, trials)
}

// Announcements

func (s *server) handleAnnouncements(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" { jsonErr(w, 405, "method not allowed"); return }
	anns, err := s.store.ListAnnouncements(true)
	if err != nil { jsonErr(w, 500, err.Error()); return }
	jsonOK(w, anns)
}
