package models

import (
    "encoding/json"
    "time"
)

type Province struct {
    ID         int    `json:"id"`
    Name       string `json:"name"`
    Code       string `json:"code"`
    TrialCount int    `json:"trial_count"`
}

type Institution struct {
    ID           int    `json:"id"`
    Name         string `json:"name"`
    ProvinceID   int    `json:"province_id"`
    ProvinceName string `json:"province_name,omitempty"`
    City         string `json:"city"`
    TrialCount   int    `json:"trial_count"`
}

type Investigator struct {
    ID              int    `json:"id"`
    Name            string `json:"name"`
    Degree          string `json:"degree,omitempty"`
    Title           string `json:"title,omitempty"`
    Phone           string `json:"phone,omitempty"`
    Email           string `json:"email,omitempty"`
    Address         string `json:"address,omitempty"`
    ZipCode         string `json:"zip_code,omitempty"`
    InstitutionID   int    `json:"institution_id"`
    InstitutionName string `json:"institution_name,omitempty"`
    TrialCount      int    `json:"trial_count"`
}

type Trial struct {
    DetailID      string          `json:"detail_id"`
    RegNo         string          `json:"reg_no"`
    Title         string          `json:"title"`
    DrugName      string          `json:"drug_name"`
    Indication    string          `json:"indication"`
    Status        string          `json:"status"`
    ApplicantName string          `json:"applicant_name"`
    Keyword       string          `json:"keyword,omitempty"`
    DetailJSON    json.RawMessage `json:"detail_json,omitempty"`
    CrawlTime     time.Time       `json:"crawl_time"`
    DataHash      string          `json:"data_hash,omitempty"`
    Published     int             `json:"published"`
    PublishedAt   *time.Time      `json:"published_at,omitempty"`
    PublishedBy   int             `json:"published_by,omitempty"`
    DiseaseZoneID int             `json:"disease_zone_id,omitempty"`
}

type TrialInstitution struct {
    TrialID          string `json:"trial_id"`
    InstitutionID    int    `json:"institution_id"`
    InvestigatorName string `json:"investigator_name"`
}

type DiseaseZone struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Keyword     string    `json:"keyword"`
    Description string    `json:"description,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}

type Announcement struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    IsPinned  int       `json:"is_pinned"`
    Published int       `json:"published"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedBy int       `json:"created_by"`
}

type CrawlLog struct {
    ID           int       `json:"id"`
    Keyword      string    `json:"keyword"`
    Pages        int       `json:"pages"`
    DiseaseZoneID int      `json:"disease_zone_id"`
    Found        int       `json:"found"`
    NewItems     int       `json:"new_items"`
    StartTime    time.Time `json:"start_time"`
    EndTime      time.Time `json:"end_time,omitempty"`
    Status       string    `json:"status"`
}

type AdminUser struct {
    ID           int       `json:"id"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`
}

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

type SearchResult struct {
    Trials   []Trial `json:"trials"`
    Total    int     `json:"total"`
    Page     int     `json:"page"`
    PageSize int     `json:"page_size"`
}

type DashboardStats struct {
    TotalTrials   int `json:"total_trials"`
    PublishedTrials int `json:"published_trials"`
    PendingTrials   int `json:"pending_trials"`
    TotalProvinces  int `json:"total_provinces"`
    TotalInstitutions int `json:"total_institutions"`
    TotalInvestigators int `json:"total_investigators"`
    TotalZones      int `json:"total_zones"`
    RecentCrawls    []CrawlLog `json:"recent_crawls"`
}
