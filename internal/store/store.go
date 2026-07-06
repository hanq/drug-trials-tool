package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"drug_trials_tool/internal/models"
)

// Store wraps the SQLite database connection.
type Store struct {
	db *sql.DB
}

// New opens or creates the SQLite database and runs migrations.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// Enable WAL mode for better concurrent access
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates all tables if they don't exist.
func (s *Store) migrate() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS provinces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			code TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS institutions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			province_id INTEGER REFERENCES provinces(id),
			city TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS investigators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			degree TEXT, title TEXT, phone TEXT,
			email TEXT, address TEXT, zip_code TEXT,
			institution_id INTEGER REFERENCES institutions(id),
			UNIQUE(name, institution_id)
		)`,
		`CREATE TABLE IF NOT EXISTS trials (
			detail_id TEXT PRIMARY KEY,
			reg_no TEXT, title TEXT, drug_name TEXT,
			indication TEXT, status TEXT, applicant_name TEXT,
			keyword TEXT,
			detail_json TEXT,
			crawl_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			data_hash TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS trial_institutions (
			trial_id TEXT REFERENCES trials(detail_id),
			institution_id INTEGER REFERENCES institutions(id),
			investigator_name TEXT,
			PRIMARY KEY(trial_id, institution_id, investigator_name)
		)`,
		`CREATE TABLE IF NOT EXISTS crawl_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			keyword TEXT, pages_crawled INTEGER DEFAULT 0,
			trials_found INTEGER DEFAULT 0, trials_new INTEGER DEFAULT 0,
			start_time TIMESTAMP, end_time TIMESTAMP,
			status TEXT DEFAULT 'running'
		)`,
	}

	for _, q := range queries {
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:40], err)
		}
	}
	return tx.Commit()
}

// ── Province ────────────────────────────────────────────────

// EnsureProvince finds or creates a province, returns its ID.
func (s *Store) EnsureProvince(name string) (int, error) {
	if name == "" {
		return 0, nil
	}
	var id int
	err := s.db.QueryRow("SELECT id FROM provinces WHERE name=?", name).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := s.db.Exec("INSERT INTO provinces(name) VALUES(?)", name)
		if err != nil {
			return 0, err
		}
		nid, _ := res.LastInsertId()
		return int(nid), nil
	}
	return id, err
}

// ListProvinces returns all provinces with trial counts.
func (s *Store) ListProvinces() ([]models.Province, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.name, p.code,
			(SELECT COUNT(DISTINCT ti.trial_id)
			 FROM trial_institutions ti
			 JOIN institutions i ON i.id = ti.institution_id
			 WHERE i.province_id = p.id) as trial_count
		FROM provinces p ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Province
	for rows.Next() {
		var p models.Province
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.TrialCount); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

// ── Institution ─────────────────────────────────────────────

// EnsureInstitution finds or creates an institution, returns its ID.
func (s *Store) EnsureInstitution(name, city string, provinceID int) (int, error) {
	if name == "" {
		return 0, nil
	}
	var id int
	err := s.db.QueryRow("SELECT id FROM institutions WHERE name=?", name).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := s.db.Exec(
			"INSERT INTO institutions(name, province_id, city) VALUES(?,?,?)",
			name, provinceID, city)
		if err != nil {
			return 0, err
		}
		nid, _ := res.LastInsertId()
		return int(nid), nil
	}
	// Update province/city if changed
	if err == nil {
		s.db.Exec("UPDATE institutions SET province_id=?, city=? WHERE id=? AND (province_id IS NULL OR province_id=0)",
			provinceID, city, id)
	}
	return id, err
}

// ListInstitutionsByProvince returns institutions in a province.
func (s *Store) ListInstitutionsByProvince(provinceID int) ([]models.Institution, error) {
	rows, err := s.db.Query(`
		SELECT i.id, i.name, i.province_id, COALESCE(p.name,''), i.city,
			(SELECT COUNT(*) FROM trial_institutions ti WHERE ti.institution_id = i.id)
		FROM institutions i
		LEFT JOIN provinces p ON p.id = i.province_id
		WHERE i.province_id = ? ORDER BY i.name`, provinceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInstitutions(rows)
}

func scanInstitutions(rows *sql.Rows) ([]models.Institution, error) {
	var result []models.Institution
	for rows.Next() {
		var inst models.Institution
		if err := rows.Scan(&inst.ID, &inst.Name, &inst.ProvinceID,
			&inst.ProvinceName, &inst.City, &inst.TrialCount); err != nil {
			return nil, err
		}
		result = append(result, inst)
	}
	return result, nil
}

// ── Investigator ────────────────────────────────────────────

// EnsureInvestigator creates or finds an investigator, returns its ID.
func (s *Store) EnsureInvestigator(name string, instID int, degree, title, phone, email, addr, zipCode string) (int, error) {
	if name == "" {
		return 0, nil
	}

	// Try update existing
	res, err := s.db.Exec(`
		INSERT INTO investigators(name, degree, title, phone, email, address, zip_code, institution_id)
		VALUES(?,?,?,?,?,?,?,?)
		ON CONFLICT(name, institution_id) DO UPDATE SET
			degree=COALESCE(NULLIF(?, ''), degree),
			title=COALESCE(NULLIF(?, ''), title),
			phone=COALESCE(NULLIF(?, ''), phone),
			email=COALESCE(NULLIF(?, ''), email),
			address=COALESCE(NULLIF(?, ''), address),
			zip_code=COALESCE(NULLIF(?, ''), zip_code)`,
		name, degree, title, phone, email, addr, zipCode, instID,
		degree, title, phone, email, addr, zipCode)

	if err != nil {
		return 0, err
	}
	nid, _ := res.LastInsertId()
	if nid == 0 {
		// Already existed, fetch id
		err = s.db.QueryRow("SELECT id FROM investigators WHERE name=? AND institution_id=?", name, instID).Scan(&nid)
		if err != nil {
			return 0, err
		}
	}
	return int(nid), nil
}

// ListInvestigatorsByInstitution returns investigators for an institution.
func (s *Store) ListInvestigatorsByInstitution(instID int) ([]models.Investigator, error) {
	rows, err := s.db.Query(`
		SELECT inv.id, inv.name, COALESCE(inv.degree,''), COALESCE(inv.title,''),
			COALESCE(inv.phone,''), COALESCE(inv.email,''), COALESCE(inv.address,''),
			COALESCE(inv.zip_code,''), inv.institution_id, COALESCE(i.name,''),
			(SELECT COUNT(*) FROM trial_institutions ti WHERE ti.investigator_name = inv.name AND ti.institution_id = inv.institution_id)
		FROM investigators inv
		JOIN institutions i ON i.id = inv.institution_id
		WHERE inv.institution_id = ? ORDER BY inv.name`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInvestigators(rows)
}

func scanInvestigators(rows *sql.Rows) ([]models.Investigator, error) {
	var result []models.Investigator
	for rows.Next() {
		var inv models.Investigator
		if err := rows.Scan(&inv.ID, &inv.Name, &inv.Degree, &inv.Title,
			&inv.Phone, &inv.Email, &inv.Address, &inv.ZipCode,
			&inv.InstitutionID, &inv.InstitutionName, &inv.TrialCount); err != nil {
			return nil, err
		}
		result = append(result, inv)
	}
	return result, nil
}

// ── Trial ───────────────────────────────────────────────────

// SaveTrial inserts or replaces a trial record. Returns whether it's new.
func (s *Store) SaveTrial(t models.Trial) (bool, error) {
	// Compute hash
	hashInput := t.RegNo + "|" + t.Title + "|" + t.DrugName + "|" + t.Indication + "|" + t.Status
	if t.DetailJSON != nil {
		hashInput += "|" + string(t.DetailJSON)
	}
	h := sha256.Sum256([]byte(hashInput))
	t.DataHash = fmt.Sprintf("%x", h[:8])

	// Check if exists with same hash
	var existingHash string
	err := s.db.QueryRow("SELECT data_hash FROM trials WHERE detail_id=?", t.DetailID).Scan(&existingHash)
	if err == nil && existingHash == t.DataHash {
		return false, nil // unchanged
	}

	detailStr := ""
	if t.DetailJSON != nil {
		detailStr = string(t.DetailJSON)
	}

	_, err = s.db.Exec(`
		INSERT INTO trials(detail_id, reg_no, title, drug_name, indication, status, applicant_name, keyword, detail_json, crawl_time, data_hash)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(detail_id) DO UPDATE SET
			reg_no=excluded.reg_no, title=excluded.title, drug_name=excluded.drug_name,
			indication=excluded.indication, status=excluded.status,
			applicant_name=excluded.applicant_name, keyword=excluded.keyword,
			detail_json=excluded.detail_json, crawl_time=excluded.crawl_time,
			data_hash=excluded.data_hash`,
		t.DetailID, t.RegNo, t.Title, t.DrugName, t.Indication, t.Status,
		t.ApplicantName, t.Keyword, detailStr, t.CrawlTime, t.DataHash)

	if err != nil {
		return false, err
	}
	return existingHash == "", nil // true if this was an insert (new)
}

// LinkTrialInstitution creates the trial-institution-investigator association.
func (s *Store) LinkTrialInstitution(trialID string, instID int, investigatorName string) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO trial_institutions(trial_id, institution_id, investigator_name)
		VALUES(?,?,?)`, trialID, instID, investigatorName)
	return err
}

// GetTrial returns a trial by detail_id.
func (s *Store) GetTrial(detailID string) (*models.Trial, error) {
	row := s.db.QueryRow(`
		SELECT detail_id, reg_no, title, drug_name, indication, status,
			COALESCE(applicant_name,''), COALESCE(keyword,''), COALESCE(detail_json,''),
			crawl_time, COALESCE(data_hash,'')
		FROM trials WHERE detail_id=?`, detailID)

	var t models.Trial
	var detailStr string
	err := row.Scan(&t.DetailID, &t.RegNo, &t.Title, &t.DrugName, &t.Indication,
		&t.Status, &t.ApplicantName, &t.Keyword, &detailStr, &t.CrawlTime, &t.DataHash)
	if err != nil {
		return nil, err
	}
	if detailStr != "" {
		t.DetailJSON = json.RawMessage(detailStr)
	}
	return &t, nil
}

// SearchTrials searches trials with filters.
func (s *Store) SearchTrials(q models.SearchQuery) (*models.SearchResult, error) {
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}

	where := "WHERE 1=1"
	var args []interface{}

	if q.Keyword != "" {
		where += " AND (t.title LIKE ? OR t.drug_name LIKE ? OR t.indication LIKE ? OR t.reg_no LIKE ?)"
		kw := "%" + q.Keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
	if q.Province != "" {
		where += " AND EXISTS (SELECT 1 FROM trial_institutions ti JOIN institutions i ON i.id=ti.institution_id JOIN provinces p ON p.id=i.province_id WHERE ti.trial_id=t.detail_id AND p.name=?)"
		args = append(args, q.Province)
	}
	if q.Institution != "" {
		where += " AND EXISTS (SELECT 1 FROM trial_institutions ti JOIN institutions i ON i.id=ti.institution_id WHERE ti.trial_id=t.detail_id AND i.name LIKE ?)"
		args = append(args, "%"+q.Institution+"%")
	}
	if q.Investigator != "" {
		where += " AND EXISTS (SELECT 1 FROM trial_institutions ti WHERE ti.trial_id=t.detail_id AND ti.investigator_name LIKE ?)"
		args = append(args, "%"+q.Investigator+"%")
	}
	if q.RegNo != "" {
		where += " AND t.reg_no LIKE ?"
		args = append(args, "%"+q.RegNo+"%")
	}
	if q.Applicant != "" {
		where += " AND t.applicant_name LIKE ?"
		args = append(args, "%"+q.Applicant+"%")
	}

	// Count
	var total int
	countQ := "SELECT COUNT(*) FROM trials t " + where
	if err := s.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Query
	offset := (q.Page - 1) * q.PageSize
	dataQ := "SELECT t.detail_id, t.reg_no, t.title, t.drug_name, t.indication, t.status, COALESCE(t.applicant_name,''), COALESCE(t.keyword,''), COALESCE(t.detail_json,''), t.crawl_time, COALESCE(t.data_hash,'') FROM trials t " + where + " ORDER BY t.crawl_time DESC LIMIT ? OFFSET ?"
	args = append(args, q.PageSize, offset)

	rows, err := s.db.Query(dataQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trials []models.Trial
	for rows.Next() {
		var t models.Trial
		var detailStr string
		if err := rows.Scan(&t.DetailID, &t.RegNo, &t.Title, &t.DrugName, &t.Indication,
			&t.Status, &t.ApplicantName, &t.Keyword, &detailStr, &t.CrawlTime, &t.DataHash); err != nil {
			return nil, err
		}
		if detailStr != "" {
			t.DetailJSON = json.RawMessage(detailStr)
		}
		trials = append(trials, t)
	}

	return &models.SearchResult{
		Trials:   trials,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

// ListTrialsByInvestigator returns trials associated with a specific investigator.
func (s *Store) ListTrialsByInvestigator(investigatorName string, institutionID int) ([]models.Trial, error) {
	rows, err := s.db.Query(`
		SELECT t.detail_id, t.reg_no, t.title, t.drug_name, t.indication, t.status,
			COALESCE(t.applicant_name,''), COALESCE(t.keyword,''), COALESCE(t.detail_json,''),
			t.crawl_time, COALESCE(t.data_hash,'')
		FROM trials t
		JOIN trial_institutions ti ON ti.trial_id = t.detail_id
		WHERE ti.investigator_name = ? AND ti.institution_id = ?
		ORDER BY t.crawl_time DESC`, investigatorName, institutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrials(rows)
}

func scanTrials(rows *sql.Rows) ([]models.Trial, error) {
	var result []models.Trial
	for rows.Next() {
		var t models.Trial
		var detailStr string
		if err := rows.Scan(&t.DetailID, &t.RegNo, &t.Title, &t.DrugName, &t.Indication,
			&t.Status, &t.ApplicantName, &t.Keyword, &detailStr, &t.CrawlTime, &t.DataHash); err != nil {
			return nil, err
		}
		if detailStr != "" {
			t.DetailJSON = json.RawMessage(detailStr)
		}
		result = append(result, t)
	}
	return result, nil
}

// ── Crawl Session ───────────────────────────────────────

// StartCrawlSession creates a new crawl session with status "running".
func (s *Store) StartCrawlSession(keyword string) (*models.CrawlSession, error) {
	res, err := s.db.Exec(
		"INSERT INTO crawl_sessions(keyword, start_time, status) VALUES(?,?,?)",
		keyword, time.Now(), "running")
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &models.CrawlSession{ID: int(id), Keyword: keyword, Status: "running"}, nil
}

// UpdateCrawlSession updates an ongoing crawl session.
func (s *Store) UpdateCrawlSession(session *models.CrawlSession) error {
	_, err := s.db.Exec(`
		UPDATE crawl_sessions SET pages_crawled=?, trials_found=?, trials_new=?,
			end_time=?, status=? WHERE id=?`,
		session.PagesCrawled, session.TrialsFound, session.TrialsNew,
		session.EndTime, session.Status, session.ID)
	return err
}

// ListCrawlSessions returns recent crawl sessions.
func (s *Store) ListCrawlSessions(limit int) ([]models.CrawlSession, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`
		SELECT id, keyword, pages_crawled, trials_found, trials_new,
			start_time, COALESCE(end_time, start_time), status
		FROM crawl_sessions ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.CrawlSession
	for rows.Next() {
		var cs models.CrawlSession
		if err := rows.Scan(&cs.ID, &cs.Keyword, &cs.PagesCrawled, &cs.TrialsFound,
			&cs.TrialsNew, &cs.StartTime, &cs.EndTime, &cs.Status); err != nil {
			return nil, err
		}
		result = append(result, cs)
	}
	return result, nil
}
// ListAllTrials returns all trials in the database.
func (s *Store) ListAllTrials() ([]models.Trial, error) {
	rows, err := s.db.Query(`
		SELECT detail_id, reg_no, title, drug_name, indication, status,
			COALESCE(applicant_name,''), COALESCE(keyword,''), COALESCE(detail_json,''),
			crawl_time, COALESCE(data_hash,'')
		FROM trials ORDER BY crawl_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrials(rows)
}

// UpdateTrial updates key fields for an existing trial.
func (s *Store) UpdateTrial(t models.Trial) error {
	detailStr := ""
	if t.DetailJSON != nil {
		detailStr = string(t.DetailJSON)
	}
	_, err := s.db.Exec(`
		UPDATE trials SET reg_no=?, title=?, drug_name=?, indication=?, status=?,
			applicant_name=?, keyword=?, detail_json=?
		WHERE detail_id=?`,
		t.RegNo, t.Title, t.DrugName, t.Indication, t.Status,
		t.ApplicantName, t.Keyword, detailStr, t.DetailID)
	return err
}
