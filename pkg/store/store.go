package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"drug_trials_tool/pkg/db"
	"drug_trials_tool/pkg/models"
)

type Store struct{ DB *sql.DB }

func New(dbPath string) (*Store, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, err
	}
	s := &Store{DB: database}
	if err := db.MigrateDB(database); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.DB.Close() }

// Province

func (s *Store) EnsureProvince(name string) (int, error) {
	if name == "" { return 0, nil }
	var id int
	err := s.DB.QueryRow("SELECT id FROM provinces WHERE name=?", name).Scan(&id)
	if err == sql.ErrNoRows {
		res, e := s.DB.Exec("INSERT INTO provinces(name) VALUES(?)", name)
		if e != nil { return 0, e }
		nid, _ := res.LastInsertId()
		return int(nid), nil
	}
	return id, err
}

func (s *Store) ListProvinces() ([]models.Province, error) {
	rows, err := s.DB.Query("SELECT p.id,p.name,COALESCE(p.code,'')," +
		"(SELECT COUNT(DISTINCT ti.trial_id) FROM trial_institutions ti " +
		"JOIN institutions i ON i.id=ti.institution_id " +
		"WHERE i.province_id=p.id AND ti.trial_id IN (SELECT detail_id FROM trials WHERE published=1)) " +
		"FROM provinces p ORDER BY p.name")
	if err != nil { return nil, err }
	defer rows.Close()
	var r []models.Province
	for rows.Next() {
		var p models.Province
		if er := rows.Scan(&p.ID, &p.Name, &p.Code, &p.TrialCount); er != nil { return nil, er }
		r = append(r, p)
	}
	return r, nil
}


// Institution

func (s *Store) EnsureInstitution(name, city string, provinceID int) (int, error) {
	if name == "" { return 0, nil }
	var id int
	err := s.DB.QueryRow("SELECT id FROM institutions WHERE name=?", name).Scan(&id)
	if err == sql.ErrNoRows {
		res, e := s.DB.Exec("INSERT INTO institutions(name,province_id,city) VALUES(?,?,?)", name, provinceID, city)
		if e != nil { return 0, e }
		nid, _ := res.LastInsertId()
		return int(nid), nil
	}
	if err == nil {
		s.DB.Exec("UPDATE institutions SET province_id=?,city=? WHERE id=? AND (province_id IS NULL OR province_id=0)", provinceID, city, id)
	}
	return id, err
}

func (s *Store) ListInstitutionsByProvince(provinceID int) ([]models.Institution, error) {
	rows, err := s.DB.Query("SELECT i.id,i.name,i.province_id,COALESCE(p.name,''),i.city,"+
		"(SELECT COUNT(*) FROM trial_institutions ti WHERE ti.institution_id=i.id "+
		"AND ti.trial_id IN (SELECT detail_id FROM trials WHERE published=1)) "+
		"FROM institutions i LEFT JOIN provinces p ON p.id=i.province_id WHERE i.province_id=? ORDER BY i.name", provinceID)
	if err != nil { return nil, err }
	defer rows.Close()
	var r []models.Institution
	for rows.Next() {
		var inst models.Institution
		if er := rows.Scan(&inst.ID, &inst.Name, &inst.ProvinceID, &inst.ProvinceName, &inst.City, &inst.TrialCount); er != nil { return nil, er }
		r = append(r, inst)
	}
	return r, nil
}

// Investigator

func (s *Store) EnsureInvestigator(name string, instID int, degree, title, phone, email, addr, zipCode string) (int, error) {
	if name == "" { return 0, nil }
	res, err := s.DB.Exec("INSERT INTO investigators(name,degree,title,phone,email,address,zip_code,institution_id) "+
		"VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(name,institution_id) DO UPDATE SET "+
		"degree=COALESCE(NULLIF(?, ''),degree), title=COALESCE(NULLIF(?, ''),title), "+
		"phone=COALESCE(NULLIF(?, ''),phone), email=COALESCE(NULLIF(?, ''),email), "+
		"address=COALESCE(NULLIF(?, ''),address), zip_code=COALESCE(NULLIF(?, ''),zip_code)",
		name, degree, title, phone, email, addr, zipCode, instID,
		degree, title, phone, email, addr, zipCode)
	if err != nil { return 0, err }
	nid, _ := res.LastInsertId()
	if nid == 0 {
		s.DB.QueryRow("SELECT id FROM investigators WHERE name=? AND institution_id=?", name, instID).Scan(&nid)
	}
	return int(nid), nil
}

func (s *Store) ListInvestigatorsByInstitution(instID int) ([]models.Investigator, error) {
	rows, err := s.DB.Query("SELECT inv.id,inv.name,COALESCE(inv.degree,''),COALESCE(inv.title,''),"+
		"COALESCE(inv.phone,''),COALESCE(inv.email,''),COALESCE(inv.address,''),"+
		"COALESCE(inv.zip_code,''),inv.institution_id,COALESCE(i.name,''),"+
		"(SELECT COUNT(*) FROM trial_institutions ti WHERE ti.investigator_name=inv.name "+
		"AND ti.institution_id=inv.institution_id "+
		"AND ti.trial_id IN (SELECT detail_id FROM trials WHERE published=1)) "+
		"FROM investigators inv JOIN institutions i ON i.id=inv.institution_id WHERE inv.institution_id=? ORDER BY inv.name", instID)
	if err != nil { return nil, err }
	defer rows.Close()
	var r []models.Investigator
	for rows.Next() {
		var inv models.Investigator
		if er := rows.Scan(&inv.ID, &inv.Name, &inv.Degree, &inv.Title, &inv.Phone, &inv.Email, &inv.Address, &inv.ZipCode, &inv.InstitutionID, &inv.InstitutionName, &inv.TrialCount); er != nil { return nil, er }
		r = append(r, inv)
	}
	return r, nil
}

// Trial

func (s *Store) SaveTrial(t models.Trial) (bool, error) {
	hashInput := t.RegNo + "|" + t.Title + "|" + t.DrugName + "|" + t.Indication + "|" + t.Status
	if t.DetailJSON != nil { hashInput += "|" + string(t.DetailJSON) }
	h := sha256.Sum256([]byte(hashInput))
	t.DataHash = fmt.Sprintf("%x", h[:8])
	var existingHash string
	err := s.DB.QueryRow("SELECT data_hash FROM trials WHERE detail_id=?", t.DetailID).Scan(&existingHash)
	if err == nil && existingHash == t.DataHash { return false, nil }
	detailStr := ""
	if t.DetailJSON != nil { detailStr = string(t.DetailJSON) }
	_, err = s.DB.Exec("INSERT INTO trials(detail_id,reg_no,title,drug_name,indication,status,"+
		"applicant_name,keyword,detail_json,crawl_time,data_hash,disease_zone_id) "+
		"VALUES(?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(detail_id) DO UPDATE SET "+
		"reg_no=excluded.reg_no,title=excluded.title,drug_name=excluded.drug_name,"+
		"indication=excluded.indication,status=excluded.status,"+
		"applicant_name=excluded.applicant_name,keyword=excluded.keyword,"+
		"detail_json=excluded.detail_json,crawl_time=excluded.crawl_time,"+
		"data_hash=excluded.data_hash",
		t.DetailID, t.RegNo, t.Title, t.DrugName, t.Indication, t.Status,
		t.ApplicantName, t.Keyword, detailStr, t.CrawlTime, t.DataHash, t.DiseaseZoneID)
	if err != nil { return false, err }
	return existingHash == "", nil
}

func (s *Store) LinkTrialInstitution(trialID string, instID int, investigatorName string) error {
	_, err := s.DB.Exec("INSERT OR IGNORE INTO trial_institutions(trial_id,institution_id,investigator_name) VALUES(?,?,?)",
		trialID, instID, investigatorName)
	return err
}

func (s *Store) GetTrial(detailID string) (*models.Trial, error) {
	row := s.DB.QueryRow("SELECT detail_id,reg_no,title,drug_name,indication,status,"+
		"COALESCE(applicant_name,''),COALESCE(keyword,''),COALESCE(detail_json,''),"+
		"crawl_time,COALESCE(data_hash,''),published,"+
		"COALESCE(published_by,0),COALESCE(disease_zone_id,0) FROM trials WHERE detail_id=?", detailID)
	var t models.Trial
	var ds string
	err := row.Scan(&t.DetailID, &t.RegNo, &t.Title, &t.DrugName, &t.Indication,
		&t.Status, &t.ApplicantName, &t.Keyword, &ds,
		&t.CrawlTime, &t.DataHash, &t.Published, &t.PublishedBy, &t.DiseaseZoneID)
	if err != nil { return nil, err }
	if ds != "" { t.DetailJSON = json.RawMessage(ds) }
	return &t, nil
}


func (s *Store) ListTrialsByInvestigator(name string, instID int) ([]models.Trial, error) {
	rows, err := s.DB.Query("SELECT t.detail_id,t.reg_no,t.title,t.drug_name,t.indication,t.status,"+
		"COALESCE(t.applicant_name,''),COALESCE(t.keyword,''),COALESCE(t.detail_json,''),"+
		"t.crawl_time,COALESCE(t.data_hash,''),t.published,"+
		"COALESCE(t.published_by,0),COALESCE(t.disease_zone_id,0) "+
		"FROM trials t JOIN trial_institutions ti ON ti.trial_id=t.detail_id "+
		"WHERE ti.investigator_name=? AND ti.institution_id=? AND t.published=1 ORDER BY t.crawl_time DESC", name, instID)
	if err != nil { return nil, err }
	defer rows.Close()
	return scanTrials(rows)
}

func scanTrials(rows *sql.Rows) ([]models.Trial, error) {
	var r []models.Trial
	for rows.Next() {
		var t models.Trial
		var ds string
		if err := rows.Scan(&t.DetailID, &t.RegNo, &t.Title, &t.DrugName, &t.Indication,
			&t.Status, &t.ApplicantName, &t.Keyword, &ds,
			&t.CrawlTime, &t.DataHash, &t.Published, &t.PublishedBy, &t.DiseaseZoneID); err != nil {
			return nil, err
		}
		if ds != "" { t.DetailJSON = json.RawMessage(ds) }
		r = append(r, t)
	}
	return r, nil
}

func (s *Store) SearchTrials(q models.SearchQuery) (*models.SearchResult, error) {
	if q.PageSize <= 0 { q.PageSize = 20 }
	if q.Page <= 0 { q.Page = 1 }
	where := "WHERE t.published=1"
	var args []interface{}
	if q.Keyword != "" {
		where += " AND (t.title LIKE ? OR t.drug_name LIKE ? OR t.indication LIKE ? OR t.reg_no LIKE ?)"
		kw := "%" + q.Keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
	if q.Province != "" {
		where += " AND EXISTS(SELECT 1 FROM trial_institutions ti JOIN institutions i ON i.id=ti.institution_id JOIN provinces p ON p.id=i.province_id WHERE ti.trial_id=t.detail_id AND p.name=?)"
		args = append(args, q.Province)
	}
	if q.Institution != "" {
		where += " AND EXISTS(SELECT 1 FROM trial_institutions ti JOIN institutions i ON i.id=ti.institution_id WHERE ti.trial_id=t.detail_id AND i.name LIKE ?)"
		args = append(args, "%"+q.Institution+"%")
	}
	if q.Investigator != "" {
		where += " AND EXISTS(SELECT 1 FROM trial_institutions ti WHERE ti.trial_id=t.detail_id AND ti.investigator_name LIKE ?)"
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
	var total int
	s.DB.QueryRow("SELECT COUNT(*) FROM trials t "+where, args...).Scan(&total)
	offset := (q.Page - 1) * q.PageSize
	sel := "SELECT t.detail_id,t.reg_no,t.title,t.drug_name,t.indication,t.status,COALESCE(t.applicant_name,''),COALESCE(t.keyword,''),COALESCE(t.detail_json,''),t.crawl_time,COALESCE(t.data_hash,''),t.published,COALESCE(t.published_by,0),COALESCE(t.disease_zone_id,0) FROM trials t " + where + " ORDER BY t.crawl_time DESC LIMIT ? OFFSET ?"
	args = append(args, q.PageSize, offset)
	rows, err := s.DB.Query(sel, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	trials, err := scanTrials(rows)
	if err != nil { return nil, err }
	return &models.SearchResult{Trials: trials, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Disease Zone

func (s *Store) ListDiseaseZones() ([]models.DiseaseZone, error) {
	rows, err := s.DB.Query("SELECT id,name,keyword,COALESCE(description,''),created_at FROM disease_zones ORDER BY name")
	if err != nil { return nil, err }
	defer rows.Close()
	var r []models.DiseaseZone
	for rows.Next() {
		var z models.DiseaseZone
		if er := rows.Scan(&z.ID, &z.Name, &z.Keyword, &z.Description, &z.CreatedAt); er != nil { return nil, er }
		r = append(r, z)
	}
	return r, nil
}

func (s *Store) CreateDiseaseZone(name, keyword, desc string) (*models.DiseaseZone, error) {
	res, err := s.DB.Exec("INSERT INTO disease_zones(name,keyword,description) VALUES(?,?,?)", name, keyword, desc)
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return &models.DiseaseZone{ID: int(id), Name: name, Keyword: keyword, Description: desc, CreatedAt: time.Now()}, nil
}

func (s *Store) DeleteDiseaseZone(id int) error {
	_, err := s.DB.Exec("DELETE FROM disease_zones WHERE id=?", id)
	return err
}

func (s *Store) ListTrialsByZone(zoneID int) ([]models.Trial, error) {
	rows, err := s.DB.Query("SELECT detail_id,reg_no,title,drug_name,indication,status,"+
		"COALESCE(applicant_name,''),COALESCE(keyword,''),COALESCE(detail_json,''),"+
		"crawl_time,COALESCE(data_hash,''),published,"+
		"COALESCE(published_by,0),COALESCE(disease_zone_id,0) "+
		"FROM trials WHERE disease_zone_id=? AND published=1 ORDER BY crawl_time DESC", zoneID)
	if err != nil { return nil, err }
	defer rows.Close()
	return scanTrials(rows)
}

// Announcement

func (s *Store) ListAnnouncements(publishedOnly bool) ([]models.Announcement, error) {
	q := "SELECT id,title,content,is_pinned,published,created_at,updated_at,COALESCE(created_by,0) FROM announcements"
	if publishedOnly { q += " WHERE published=1" }
	q += " ORDER BY is_pinned DESC, created_at DESC"
	rows, err := s.DB.Query(q)
	if err != nil { return nil, err }
	defer rows.Close()
	var r []models.Announcement
	for rows.Next() {
		var a models.Announcement
		if er := rows.Scan(&a.ID, &a.Title, &a.Content, &a.IsPinned, &a.Published, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy); er != nil { return nil, er }
		r = append(r, a)
	}
	return r, nil
}

func (s *Store) CreateAnnouncement(title, content string, createdBy int) (*models.Announcement, error) {
	now := time.Now()
	res, err := s.DB.Exec("INSERT INTO announcements(title,content,created_by,created_at,updated_at) VALUES(?,?,?,?,?)",
		title, content, createdBy, now, now)
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return &models.Announcement{ID: int(id), Title: title, Content: content, IsPinned: 0, Published: 1, CreatedAt: now, UpdatedAt: now, CreatedBy: createdBy}, nil
}

func (s *Store) UpdateAnnouncement(id int, title, content string) error {
	_, err := s.DB.Exec("UPDATE announcements SET title=?,content=?,updated_at=? WHERE id=?", title, content, time.Now(), id)
	return err
}

func (s *Store) DeleteAnnouncement(id int) error {
	_, err := s.DB.Exec("DELETE FROM announcements WHERE id=?", id)
	return err
}

// Admin

func (s *Store) CreateAdmin(username, passwordHash string) error {
	_, err := s.DB.Exec("INSERT INTO admin_users(username,password_hash) VALUES(?,?)", username, passwordHash)
	return err
}

func (s *Store) GetAdminByUsername(username string) (*models.AdminUser, error) {
	row := s.DB.QueryRow("SELECT id,username,password_hash,created_at FROM admin_users WHERE username=?", username)
	var u models.AdminUser
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil { return nil, err }
	return &u, nil
}

// Publish

func (s *Store) PublishTrial(detailID string, adminID int) error {
	_, err := s.DB.Exec("UPDATE trials SET published=1,published_at=?,published_by=? WHERE detail_id=?", time.Now(), adminID, detailID)
	return err
}

func (s *Store) BatchPublish(zoneID int, keyword string) (int, error) {
	var res sql.Result
	var err error
	now := time.Now()
	if zoneID > 0 && keyword != "" {
		res, err = s.DB.Exec("UPDATE trials SET published=1,published_at=? WHERE disease_zone_id=? AND keyword=? AND published=0", now, zoneID, keyword)
	} else if zoneID > 0 {
		res, err = s.DB.Exec("UPDATE trials SET published=1,published_at=? WHERE disease_zone_id=? AND published=0", now, zoneID)
	} else if keyword != "" {
		res, err = s.DB.Exec("UPDATE trials SET published=1,published_at=? WHERE keyword=? AND published=0", now, keyword)
	} else {
		return 0, nil
	}
	if err != nil { return 0, err }
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (s *Store) ListUnpublishedTrials(zoneID int, keyword string) ([]models.Trial, error) {
	where := "WHERE t.published=0"
	var args []interface{}
	if zoneID > 0 { where += " AND t.disease_zone_id=?"; args = append(args, zoneID) }
	if keyword != "" { where += " AND t.keyword=?"; args = append(args, keyword) }
	rows, err := s.DB.Query("SELECT t.detail_id,t.reg_no,t.title,t.drug_name,t.indication,t.status,"+
		"COALESCE(t.applicant_name,''),COALESCE(t.keyword,''),COALESCE(t.detail_json,''),"+
		"t.crawl_time,COALESCE(t.data_hash,''),t.published,"+
		"COALESCE(t.published_by,0),COALESCE(t.disease_zone_id,0) "+
		"FROM trials t "+where+" ORDER BY t.crawl_time DESC", args...)
	if err != nil { return nil, err }
	defer rows.Close()
	return scanTrials(rows)
}

// Crawl Log

func (s *Store) StartCrawlLog(keyword string, pages, zoneID int) (*models.CrawlLog, error) {
	res, err := s.DB.Exec("INSERT INTO crawl_logs(keyword,pages,disease_zone_id,start_time,status) VALUES(?,?,?,?,?)",
		keyword, pages, zoneID, time.Now(), "running")
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return &models.CrawlLog{ID: int(id), Keyword: keyword, Pages: pages, DiseaseZoneID: zoneID, Status: "running"}, nil
}

func (s *Store) FinishCrawlLog(id, found, newItems int, status string) error {
	_, err := s.DB.Exec("UPDATE crawl_logs SET found=?,new_items=?,end_time=?,status=? WHERE id=?",
		found, newItems, time.Now(), status, id)
	return err
}

func (s *Store) ListCrawlLogs(limit int) ([]models.CrawlLog, error) {
	if limit <= 0 { limit = 20 }
	rows, err := s.DB.Query("SELECT id,keyword,pages,COALESCE(disease_zone_id,0),found,new_items,"+
		"start_time,COALESCE(end_time,start_time),status FROM crawl_logs ORDER BY id DESC LIMIT ?", limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var r []models.CrawlLog
	for rows.Next() {
		var l models.CrawlLog
		if er := rows.Scan(&l.ID, &l.Keyword, &l.Pages, &l.DiseaseZoneID, &l.Found, &l.NewItems, &l.StartTime, &l.EndTime, &l.Status); er != nil { return nil, er }
		r = append(r, l)
	}
	return r, nil
}

// Dashboard

func (s *Store) DashboardStats() (*models.DashboardStats, error) {
	st := &models.DashboardStats{}
	s.DB.QueryRow("SELECT COUNT(*) FROM trials").Scan(&st.TotalTrials)
	s.DB.QueryRow("SELECT COUNT(*) FROM trials WHERE published=1").Scan(&st.PublishedTrials)
	s.DB.QueryRow("SELECT COUNT(*) FROM trials WHERE published=0").Scan(&st.PendingTrials)
	s.DB.QueryRow("SELECT COUNT(*) FROM provinces").Scan(&st.TotalProvinces)
	s.DB.QueryRow("SELECT COUNT(*) FROM institutions").Scan(&st.TotalInstitutions)
	s.DB.QueryRow("SELECT COUNT(*) FROM investigators").Scan(&st.TotalInvestigators)
	s.DB.QueryRow("SELECT COUNT(*) FROM disease_zones").Scan(&st.TotalZones)
	logs, err := s.ListCrawlLogs(5)
	if err == nil { st.RecentCrawls = logs }
	return st, nil
}
