package models

import (
	"encoding/json"
	"time"
)

// Province represents a Chinese province/region with trial counts.
type Province struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	TrialCount   int    `json:"trial_count"`
}

// Institution represents a research institution.
type Institution struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ProvinceID   int    `json:"province_id"`
	ProvinceName string `json:"province_name,omitempty"`
	City         string `json:"city"`
	TrialCount   int    `json:"trial_count"`
}

// Investigator represents a principal investigator.
type Investigator struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Degree         string `json:"degree,omitempty"`
	Title          string `json:"title,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Email          string `json:"email,omitempty"`
	Address        string `json:"address,omitempty"`
	ZipCode        string `json:"zip_code,omitempty"`
	InstitutionID  int    `json:"institution_id"`
	InstitutionName string `json:"institution_name,omitempty"`
	TrialCount     int    `json:"trial_count"`
}

// Trial represents a clinical trial with key fields for query.
type Trial struct {
	DetailID      string `json:"detail_id"`      // 32-hex primary key
	RegNo         string `json:"reg_no"`          // 登记号
	Title         string `json:"title"`           // 试验通俗题目
	DrugName      string `json:"drug_name"`       // 药物名称
	Indication    string `json:"indication"`      // 适应症
	Status        string `json:"status"`          // 试验状态
	ApplicantName string `json:"applicant_name"`  // 申请人名称
	Keyword       string `json:"keyword"`         // 搜索关键词

	// Full detail JSON — stores all 7 sections parsed from collapseTwo
	DetailJSON json.RawMessage `json:"detail_json,omitempty"`

	CrawlTime   time.Time `json:"crawl_time"`
	DataHash    string    `json:"data_hash"`
}

// CrawlSession tracks one crawl execution.
type CrawlSession struct {
	ID           int       `json:"id"`
	Keyword      string    `json:"keyword"`
	PagesCrawled int       `json:"pages_crawled"`
	TrialsFound  int       `json:"trials_found"`
	TrialsNew    int       `json:"trials_new"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"` // running, completed, failed
}

// TrialDetail stores the full parsed structure of all 7 sections.
type TrialDetail struct {
	DetailID string           `json:"detail_id"`
	RegNo    string           `json:"reg_no"`
	Sections []DetailSection  `json:"sections"`
}

// DetailSection represents one section of the detail page (e.g., "一、题目和背景信息").
type DetailSection struct {
	Title  string        `json:"title"`
	Fields []DetailField `json:"fields,omitempty"`
	Tables []DetailTable `json:"tables,omitempty"`
}

// DetailField represents a key-value field pair in a table.
type DetailField struct {
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

// DetailTable represents a table with a header row and data rows.
type DetailTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// InstitutionRow is one row from "各参加机构信息" table.
type InstitutionRow struct {
	Index        int    `json:"index"`
	Institution  string `json:"institution"`
	Investigator string `json:"investigator"`
	Country      string `json:"country"`
	Province     string `json:"province"`
	City         string `json:"city"`
}

// SearchQuery represents frontend search/filter parameters.
type SearchQuery struct {
	Keyword      string `json:"keyword"`
	Province     string `json:"province"`
	Institution  string `json:"institution"`
	Investigator string `json:"investigator"`
	RegNo        string `json:"reg_no"`
	Applicant    string `json:"applicant"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

// SearchResult wraps paginated search results.
type SearchResult struct {
	Trials     []Trial `json:"trials"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
}

// APIResponse is the standard API response wrapper.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
