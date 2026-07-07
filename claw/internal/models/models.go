package models


type TrialInfo struct {
	Index      int
	Title      string
	DrugName   string
	Indication string
	Status     string
	RegNo      string
	DetailID   string
}

type CrawlResult struct {
	Detail       *TrialDetail
	Institutions []InstitutionRow
	Investigator *Investigator
}

type Investigator struct {
	Name            string
	Degree          string
	Title           string
	Phone           string
	Email           string
	Address         string
	ZipCode         string
	InstitutionName string
}

type InstitutionRow struct {
	Index        int
	Institution  string
	Investigator string
	Country      string
	Province     string
	City         string
}

type TrialDetail struct {
	DetailID string
	RegNo    string
	Sections []DetailSection
}

type DetailSection struct {
	Title  string
	Fields []DetailField
	Tables []DetailTable
}

type DetailField struct {
	Label  string
	Values []string
}

type DetailTable struct {
	Headers []string
	Rows    [][]string
}
