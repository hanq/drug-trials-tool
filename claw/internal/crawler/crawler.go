package crawler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"drug_trials_tool/claw/internal/client"
	"drug_trials_tool/pkg/models"
	clawmodels "drug_trials_tool/claw/internal/models"
)

type Crawler struct {
	cl       *client.Client
	lastHTML string
}

func New(cl *client.Client) *Crawler {
	return &Crawler{cl: cl}
}

func (cr *Crawler) GetLastSearchHTML() (string, bool) {
	return cr.lastHTML, cr.lastHTML != ""
}

// ── Search (list page) ─────────────────────────────────────

func (cr *Crawler) searchPage(keyword string, page int) ([]TrialInfo, int, error) {
	params := map[string]string{
		"keywords": keyword, "currentpage": strconv.Itoa(page),
		"sort": "desc", "sort2": "", "rule": "CTR", "secondLevel": "0",
		"id": "", "ckm_index": "",
	}
	status, body, err := cr.cl.Post("/clinicaltrials.searchlist.dhtml", params)
	if err != nil {
		return nil, 0, fmt.Errorf("search request page %d: %w", page, err)
	}
	if status != 200 {
		return nil, 0, fmt.Errorf("search returned status %d", status)
	}
	bodyStr := string(body)
	if containsWAFChallenge(bodyStr) {
		fmt.Printf("  WAF challenge on page %d, retrying...\n", page)
		if err := cr.cl.InitSession(); err != nil {
			return nil, 0, fmt.Errorf("re-init session: %w", err)
		}
		status, body, err = cr.cl.Post("/clinicaltrials.searchlist.dhtml", params)
		if err != nil {
			return nil, 0, fmt.Errorf("search retry page %d: %w", page, err)
		}
		bodyStr = string(body)
	}   // closes WAF challenge if
	cr.lastHTML = bodyStr
	return parseSearchResults(bodyStr), extractTotalPages(bodyStr), nil
}

func (cr *Crawler) SearchByDisease(keyword string) ([]TrialInfo, error) {
	fmt.Printf("Searching for disease: %s\n", keyword)
	trials, _, err := cr.searchPage(keyword, 1)
	if err != nil {
		return nil, err
	}
	fmt.Printf("  Found %d trials on page 1\n", len(trials))
	return trials, nil
}

func (cr *Crawler) SearchAllPages(keyword string, maxPages int) ([]TrialInfo, error) {
	fmt.Printf("Searching for disease: %s (all pages)\n", keyword)
	var allTrials []TrialInfo
	seenIDs := make(map[string]bool)
	page1Trials, tp, err := cr.searchPage(keyword, 1)
	if err != nil {
		return nil, fmt.Errorf("page 1: %w", err)
	}
	totalPages := tp
	fmt.Printf("  Page 1: %d trials, total pages: %d\n", len(page1Trials), totalPages)
	if maxPages > 0 && maxPages < totalPages {
		totalPages = maxPages
	}
	for _, t := range page1Trials {
		if !seenIDs[t.DetailID] {
			seenIDs[t.DetailID] = true
			allTrials = append(allTrials, t)
		}
	}
	for p := 2; p <= totalPages; p++ {
		trials, _, err := cr.searchPage(keyword, p)
		if err != nil {
			fmt.Printf("  Page %d error: %v\n", p, err)
			break
		}
		if len(trials) == 0 {
			break
		}
		added := 0
		for _, t := range trials {
			if !seenIDs[t.DetailID] {
				seenIDs[t.DetailID] = true
				allTrials = append(allTrials, t)
				added++
			}
		}
		fmt.Printf("  Page %d: %d trials (+%d new)\n", p, len(trials), added)
	}
	fmt.Printf("  Total: %d unique trials across %d pages\n", len(allTrials), totalPages)
	return allTrials, nil
}

// ── Detail page ────────────────────────────────────────────

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
	Trial        models.Trial
	Detail       *clawmodels.TrialDetail
	Institutions []clawmodels.InstitutionRow
	Investigator *clawmodels.Investigator
}

func (cr *Crawler) FetchAndParseDetail(detailID, regNo, keyword string, pageDir string) (*CrawlResult, error) {
	path := "/clinicaltrials.searchlistdetail.dhtml?id=" + detailID
	html, err := cr.GetPageByURL(path)
	if err != nil {
		return nil, fmt.Errorf("fetch detail %s: %w", detailID, err)
	}
	if pageDir != "" {
		os.MkdirAll(pageDir, 0755)
		os.WriteFile(filepath.Join(pageDir, detailID+".html"), []byte(html), 0644)
	}

	detail := ParseTrialDetail(html, detailID)
	institutions := extractInstitutionRows(detail, detailID)
	inv := extractMainInvestigator(detail)
	detailJSON, _ := json.Marshal(detail)

	t := models.Trial{
		DetailID:   detailID,
		RegNo:      regNo,
		Keyword:    keyword,
		DetailJSON: detailJSON,
		CrawlTime:  time.Now(),
	}

	// Extract key fields from parsed sections
	for _, sec := range detail.Sections {
		for _, f := range sec.Fields {
			v := ""
			if len(f.Values) > 0 {
				v = f.Values[0]
			}
			switch sec.Title {
			case "一、题目和背景信息":
				switch f.Label {
				case "登记号":
					if t.RegNo == "" { t.RegNo = v }
				case "药物名称":
					if t.DrugName == "" { t.DrugName = v }
				case "适应症":
					if t.Indication == "" { t.Indication = v }
				case "试验通俗题目":
					if t.Title == "" { t.Title = v }
				case "试验专业题目":
					if t.Title == "" { t.Title = v }
				}
			case "二、申请人信息":
				if f.Label == "申请人名称" && t.ApplicantName == "" {
					t.ApplicantName = v
				}
			case "六、试验状态信息":
				if strings.Contains(f.Label, "试验状态") || f.Label == "1、试验状态" {
					t.Status = v
				}
			}
		}
	}
	// Fallback status: look for text in status section
	if t.Status == "" {
		for _, sec := range detail.Sections {
			if strings.Contains(sec.Title, "试验状态") {
				for _, f := range sec.Fields {
					if f.Label == "1、试验状态" && len(f.Values) > 0 {
						t.Status = f.Values[0]
					}
				}
			}
		}
	}

	return &CrawlResult{
		Trial: t, Detail: detail,
		Institutions: institutions, Investigator: inv,
	}, nil
}

// ── ParseTrialDetail: extracts all 7 sections from the full HTML ──

func ParseTrialDetail(html string, detailID string) *clawmodels.TrialDetail {
	detail := &clawmodels.TrialDetail{DetailID: detailID}

	// Use strings.Index to find section title divs reliably
	marker := `<div class="searchDetailPartTit"`
	pos := 0

	for {
		idx := strings.Index(html[pos:], marker)
		if idx < 0 {
			break
		}
		start := pos + idx

		// Extract the title text between <div ...> and </div>
		closeTag := strings.Index(html[start:], "</div>")
		if closeTag < 0 {
			break
		}
		titleBlock := html[start : start+closeTag+6]
		title := strings.TrimSpace(stripTags(titleBlock))
		titleEnd := start + closeTag + 6 // position right after </div>

		sec := clawmodels.DetailSection{Title: title}

		// Content: everything from after the title div to the next title or end
		contentStart := titleEnd

		// Find next marker position
		nextMarker := strings.Index(html[contentStart:], marker)
		var content string
		if nextMarker >= 0 {
			content = html[contentStart : contentStart+nextMarker]
		} else {
			content = html[contentStart:]
		}

		// Parse tables in content
		tableRe := regexp.MustCompile(`(?s)<table[^>]*>(.*?)</table>`)
		for _, tm := range tableRe.FindAllStringSubmatch(content, -1) {
			parseTable(tm[1], &sec)
		}

		// Parse sDPTit2 sub-sections (text blocks between sub-titles)
		subMarker := `<div class="sDPTit2"`
		subPos := 0
		for {
			si := strings.Index(content[subPos:], subMarker)
			if si < 0 {
				break
			}
			sStart := subPos + si
			sClose := strings.Index(content[sStart:], "</div>")
			if sClose < 0 {
				break
			}
			subTitle := strings.TrimSpace(stripTags(content[sStart : sStart+sClose+6]))
			afterDiv := sStart + sClose + 6
			nextSub := strings.Index(content[afterDiv:], subMarker)
			var subContent string
			if nextSub >= 0 {
				subContent = content[afterDiv : afterDiv+nextSub]
			} else {
				subContent = content[afterDiv:]
			}
			stripped := strings.TrimSpace(stripTags(subContent))
			if stripped != "" {
				sec.Fields = append(sec.Fields, clawmodels.DetailField{
					Label: subTitle, Values: []string{stripped},
				})
			}
			subPos = sStart + sClose + 6
		}

		detail.Sections = append(detail.Sections, sec)
		pos = titleEnd
	}

	return detail
}

// ── Table parsing ─────────────────────────────────────────

func parseTable(tableHTML string, sec *clawmodels.DetailSection) {
	rowRe := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	rows := rowRe.FindAllStringSubmatch(tableHTML, -1)
	if len(rows) == 0 {
		return
	}

	firstRow := rows[0][1]
	hasHeader := strings.Contains(firstRow, "<th")

	if hasHeader {
		thRe := regexp.MustCompile(`(?s)<th[^>]*>(.*?)</th>`)
		hdrs := thRe.FindAllStringSubmatch(firstRow, -1)
		var headers []string
		for _, h := range hdrs {
			headers = append(headers, strings.TrimSpace(stripTags(h[1])))
		}
		if len(headers) >= 2 && headers[0] == "序号" {
			// Data table: extract as rows
			var dataRows [][]string
			for _, row := range rows[1:] {
				tdRe := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
				tdMatches := tdRe.FindAllStringSubmatch(row[1], -1)
				var rowData []string
				for _, td := range tdMatches {
					rowData = append(rowData, strings.TrimSpace(stripTags(td[1])))
				}
				if len(rowData) > 0 {
					dataRows = append(dataRows, rowData)
				}
			}
			sec.Tables = append(sec.Tables, clawmodels.DetailTable{
				Headers: headers, Rows: dataRows,
			})
		} else {
			// Label-value rows with <th> + <td>
			for _, row := range rows {
				thRe := regexp.MustCompile(`(?s)<th[^>]*>(.*?)</th>`)
				tdRe := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
				thm := thRe.FindAllStringSubmatch(row[1], -1)
				tdm := tdRe.FindAllStringSubmatch(row[1], -1)
				var labels, values []string
				for _, m := range thm {
					labels = append(labels, strings.TrimSpace(stripTags(m[1])))
				}
				for _, m := range tdm {
					values = append(values, strings.TrimSpace(stripTags(m[1])))
				}
				if len(labels) == 1 && len(values) >= 1 {
					sec.Fields = append(sec.Fields, clawmodels.DetailField{
						Label: labels[0], Values: values,
					})
				} else if len(labels) > 0 && len(labels) == len(values) {
					for i := range labels {
						sec.Fields = append(sec.Fields, clawmodels.DetailField{
							Label: labels[i], Values: []string{values[i]},
						})
					}
				}
			}
		}
	} else {
		// All rows are label-value
		for _, row := range rows {
			thRe := regexp.MustCompile(`(?s)<th[^>]*>(.*?)</th>`)
			tdRe := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
			thm := thRe.FindAllStringSubmatch(row[1], -1)
			tdm := tdRe.FindAllStringSubmatch(row[1], -1)
			var labels, values []string
			for _, m := range thm {
				labels = append(labels, strings.TrimSpace(stripTags(m[1])))
			}
			for _, m := range tdm {
				values = append(values, strings.TrimSpace(stripTags(m[1])))
			}
			if len(labels) == 1 && len(values) >= 1 {
				sec.Fields = append(sec.Fields, clawmodels.DetailField{
					Label: labels[0], Values: values,
				})
			} else if len(labels) > 0 && len(labels) == len(values) {
				for i := range labels {
					sec.Fields = append(sec.Fields, clawmodels.DetailField{
						Label: labels[i], Values: []string{values[i]},
					})
				}
			}
		}
	}
}

// ── Extract institution/investigator data ──────────────────

func extractInstitutionRows(detail *clawmodels.TrialDetail, detailID string) []clawmodels.InstitutionRow {
	var result []clawmodels.InstitutionRow
	for _, sec := range detail.Sections {
		if !strings.Contains(sec.Title, "研究者信息") {
			continue
		}
		for _, tb := range sec.Tables {
			if len(tb.Headers) < 3 || !strings.Contains(tb.Headers[1], "机构") {
				continue
			}
			for _, row := range tb.Rows {
				ir := clawmodels.InstitutionRow{Index: len(result) + 1}
				if len(row) >= 2 { ir.Institution = row[1] }
				if len(row) >= 3 { ir.Investigator = row[2] }
				if len(row) >= 4 { ir.Country = row[3] }
				if len(row) >= 5 { ir.Province = row[4] }
				if len(row) >= 6 { ir.City = row[5] }
				result = append(result, ir)
			}
		}
	}
	return result
}

func extractMainInvestigator(detail *clawmodels.TrialDetail) *clawmodels.Investigator {
	for _, sec := range detail.Sections {
		if !strings.Contains(sec.Title, "研究者信息") {
			continue
		}
		inv := &clawmodels.Investigator{}
		found := false
		for _, f := range sec.Fields {
			v := ""
			if len(f.Values) > 0 { v = f.Values[0] }
			switch f.Label {
			case "姓名": inv.Name = v; found = true
			case "学位": inv.Degree = v
			case "职称": inv.Title = v
			case "电话": inv.Phone = v
			case "Email": inv.Email = v
			case "邮政地址": inv.Address = v
			case "邮编": inv.ZipCode = v
			case "单位名称": inv.InstitutionName = v
			}
		}
		if inv.InstitutionName == "" {
			for _, tb := range sec.Tables {
				if len(tb.Rows) > 0 && len(tb.Rows[0]) >= 2 {
					inv.InstitutionName = tb.Rows[0][1]
				}
			}
		}
		if found {
			return inv
		}
	}
	return nil
}

// ── Utility ────────────────────────────────────────────────

func (cr *Crawler) GetPageByURL(path string) (string, error) {
	status, body, err := cr.cl.Get(path)
	if err != nil {
		return "", fmt.Errorf("fetch page %s: %w", path, err)
	}
	bodyStr := string(body)
	if status != 200 || containsWAFChallenge(bodyStr) {
		if status == 200 {
			fmt.Printf("  WAF content in response, retrying...\n")
			if err := cr.cl.InitSession(); err != nil {
				return "", fmt.Errorf("re-init session: %w", err)
			}
			status, body, err = cr.cl.Get(path)
			if err != nil {
				return "", fmt.Errorf("retry: %w", err)
			}
			bodyStr = string(body)
			if status != 200 || containsWAFChallenge(bodyStr) {
				return "", fmt.Errorf("page %s WAF persisted", path)
			}
		} else {
			return "", fmt.Errorf("page %s returned status %d", path, status)
		}
	}
	return bodyStr, nil
}

func extractTotalPages(html string) int {
	re := regexp.MustCompile(`共\s*<i>\s*(\d+)\s*</i>\s*页`)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		page, _ := strconv.Atoi(m[1])
		if page > 0 { return page }
	}
	re2 := regexp.MustCompile(`(?i)onclick\s*=\s*["']gotopage\s*\(\s*(\d+)\s*\)`)
	matches := re2.FindAllStringSubmatch(html, -1)
	maxPage := 0
	for _, m := range matches {
		page, _ := strconv.Atoi(m[1])
		if page > maxPage { maxPage = page }
	}
	if maxPage > 0 { return maxPage }
	return 1
}

func parseSearchResults(html string) []TrialInfo {
	var trials []TrialInfo
	seen := make(map[string]bool)
	rowRe := regexp.MustCompile(`(?s)<tr[^>]*>\s*(.*?)\s*</tr>`)
	rows := rowRe.FindAllStringSubmatch(html, -1)
	for _, row := range rows {
		rc := row[1]
		if strings.Contains(rc, "<th") { continue }
		aRe := regexp.MustCompile(`(?s)<a\s+[^>]*id="([a-f0-9]{32}|[0-9]+)(?:_[a-z0-9]+)?"[^>]*>\s*(.*?)\s*</a>`)
		aTags := aRe.FindAllStringSubmatch(rc, -1)
		if len(aTags) == 0 { continue }
		did := aTags[0][1]
		if seen[did] { continue }
		seen[did] = true
		t := TrialInfo{Index: len(trials) + 1, DetailID: did}
		if len(aTags) >= 1 { t.RegNo = strings.TrimSpace(aTags[0][2]) }
		if len(aTags) >= 2 { t.Status = strings.TrimSpace(aTags[1][2]) }
		if len(aTags) >= 3 { t.DrugName = strings.TrimSpace(aTags[2][2]) }
		if len(aTags) >= 4 { t.Indication = strings.TrimSpace(aTags[3][2]) }
		if len(aTags) >= 5 { t.Title = strings.TrimSpace(aTags[4][2]) }
		if t.Title == "" { t.Title = t.DrugName }
		if t.Title == "" { t.Title = t.RegNo }
		trials = append(trials, t)
	}
	return trials
}

func containsWAFChallenge(body string) bool {
	// Real content page: contains the database tables with this class
	if strings.Contains(body, "searchDetailTable") || strings.Contains(body, "collapseTwo") {
		return false
	}
	// WAF challenge page: small page with obfuscated JS
	return len(body) < 10000 && (strings.Contains(body, "FSSBBIl1UgzbN7N80S") ||
		strings.Contains(body, "document.createElement"))
}

func stripTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}


