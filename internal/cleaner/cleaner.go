package cleaner

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"drug_trials_tool/internal/models"
	"drug_trials_tool/internal/store"
)

// Cleaner handles post-crawl data cleaning.
type Cleaner struct {
	store   *store.Store
	pageDir string
}

// New creates a new Cleaner.
func New(s *store.Store, pageDir string) *Cleaner {
	return &Cleaner{store: s, pageDir: pageDir}
}

// CleanAll runs all cleaning steps on existing data.
func (c *Cleaner) CleanAll() error {
	fmt.Println("=== Data Cleaning ===")

	if err := c.cleanTextFields(); err != nil {
		return fmt.Errorf("clean text: %w", err)
	}
	if err := c.fixApplicantNames(); err != nil {
		return fmt.Errorf("fix applicant: %w", err)
	}
	if err := c.normalizeInstitutions(); err != nil {
		return fmt.Errorf("normalize institutions: %w", err)
	}

	fmt.Println("=== Cleaning complete ===")
	return nil
}

// cleanTextFields trims whitespace and decodes HTML entities in key fields.
func (c *Cleaner) cleanTextFields() error {
	fmt.Println("  Cleaning text fields...")

	// Get all trials
	trials, err := c.store.ListAllTrials()
	if err != nil {
		return err
	}

	updated := 0
	for _, t := range trials {
		changed := false
		orig := t

		t.Title = cleanText(t.Title)
		t.DrugName = cleanText(t.DrugName)
		t.Indication = cleanText(t.Indication)
		t.Status = cleanText(t.Status)
		t.ApplicantName = cleanText(t.ApplicantName)

		if t.Title != orig.Title || t.DrugName != orig.DrugName ||
			t.Indication != orig.Indication || t.Status != orig.Status ||
			t.ApplicantName != orig.ApplicantName {
			changed = true
		}

		// Also clean detail JSON if present
		if t.DetailJSON != nil {
			var detail models.TrialDetail
			if err := json.Unmarshal(t.DetailJSON, &detail); err == nil {
				cleaned := cleanDetailJSON(&detail)
				if cleaned {
					newJSON, _ := json.Marshal(detail)
					t.DetailJSON = newJSON
					changed = true
				}
			}
		}

		if changed {
			c.store.UpdateTrial(t)
			updated++
		}
	}
	fmt.Printf("    Updated %d / %d trials\n", updated, len(trials))
	return nil
}

// fixApplicantNames re-extracts applicant names from saved HTML pages.
func (c *Cleaner) fixApplicantNames() error {
	fmt.Println("  Fixing applicant names from saved HTML...")

	trials, err := c.store.ListAllTrials()
	if err != nil {
		return err
	}

	updated := 0
	for _, t := range trials {
		htmlPath := filepath.Join(c.pageDir, t.DetailID+".html")
		data, err := os.ReadFile(htmlPath)
		if err != nil {
			continue
		}

		name := extractApplicantName(string(data))
		if name != "" && name != t.ApplicantName {
			t.ApplicantName = name
			c.store.UpdateTrial(t)
			updated++
		}
	}
	fmt.Printf("    Fixed %d applicant names\n", updated)
	return nil
}

// normalizeInstitutions cleans institution names.
func (c *Cleaner) normalizeInstitutions() error {
	fmt.Println("  Normalizing institutions...")
	// This is a lightweight pass - major cleanup is in the API layer
	return nil
}

// ── Helpers ────────────────────────────────────────────────

// cleanText trims whitespace and decodes HTML entities.
func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = html.UnescapeString(s)
	// Collapse multiple whitespace
	spaceRe := regexp.MustCompile(`\s+`)
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// cleanDetailJSON recursively cleans text in all fields of the detail JSON.
func cleanDetailJSON(detail *models.TrialDetail) bool {
	changed := false
	for i, sec := range detail.Sections {
		sec.Title = cleanText(sec.Title)
		if sec.Title != detail.Sections[i].Title {
			changed = true
			detail.Sections[i].Title = sec.Title
		}
		for j, f := range sec.Fields {
			f.Label = cleanText(f.Label)
			if f.Label != sec.Fields[j].Label {
				changed = true
			}
			for k, v := range f.Values {
				f.Values[k] = cleanText(v)
				if f.Values[k] != sec.Fields[j].Values[k] {
					changed = true
				}
			}
			detail.Sections[i].Fields[j] = f
		}
		// Clean tables
		for j, tb := range sec.Tables {
			for k, h := range tb.Headers {
				tb.Headers[k] = cleanText(h)
			}
			for k, row := range tb.Rows {
				for r, v := range row {
					tb.Rows[k][r] = cleanText(v)
				}
			}
			detail.Sections[i].Tables[j] = tb
		}
	}
	return changed
}

// extractApplicantName extracts the applicant/company name from the detail page HTML.
// The HTML structure is:
//   <div class="searchDetailPartTit">二、申请人信息</div>
//   <table ...>
//     <tr>
//       <th>申请人名称</th>
//       <td colspan="5">
//         <div class="input-group">
//           <span class="input-group-addon">1</span>
//           <input type="text" ... value="Company Name">
//         </div>
//       </td>
//     </tr>
//   </table>
func extractApplicantName(html string) string {
	// Find the 申请人信息 section
	marker := `<div class="searchDetailPartTit">二、申请人信息</div>`
	idx := strings.Index(html, marker)
	if idx < 0 {
		return ""
	}
	section := html[idx:]

	// Find the applicant name table row
	rowMarker := `<th>申请人名称</th>`
	rowIdx := strings.Index(section, rowMarker)
	if rowIdx < 0 {
		return ""
	}
	rowContent := section[rowIdx:]

	// Find all input values
	re := regexp.MustCompile(`<input[^>]*value="([^"]*)"`)
	matches := re.FindAllStringSubmatch(rowContent, -1)

	var names []string
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, "; ")
}
