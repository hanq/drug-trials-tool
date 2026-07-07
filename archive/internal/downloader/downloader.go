package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"drug_trials_tool/internal/client"
)

// Downloader handles downloading DOC documents from trial detail pages.
type Downloader struct {
	cl        *client.Client
	outputDir string
}

// New creates a new Downloader.
func New(cl *client.Client, outputDir string) *Downloader {
	return &Downloader{
		cl:        cl,
		outputDir: outputDir,
	}
}

// Result represents the outcome of a single download attempt.
type Result struct {
	Title    string
	ID       string
	FilePath string
	Success  bool
	Error    string
}

// DownloadByID downloads a DOC document by its download ID.
// POST /clinicaltrials.searchlistdetail.dhtml?_export=doc with id=XXXXX
func (d *Downloader) DownloadByID(id, title string) (*Result, error) {
	params := map[string]string{
		"id": id,
	}

	status, body, err := d.cl.PostDownload(
		"/clinicaltrials.searchlistdetail.dhtml?_export=doc",
		params,
	)
	if err != nil {
		return &Result{
			ID:      id,
			Title:   title,
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}

	if status != 200 {
		return &Result{
			ID:      id,
			Title:   title,
			Success: false,
			Error:   fmt.Sprintf("download returned status %d", status),
		}, nil
	}

	ext := detectFileExt(body, title)
	filename := sanitizeFilename(title)
	if filename == "" {
		filename = id
	}
	filename += ext

	filePath := filepath.Join(d.outputDir, filename)

	if err := os.MkdirAll(d.outputDir, 0755); err != nil {
		return &Result{
			ID:      id,
			Title:   title,
			Success: false,
			Error:   fmt.Sprintf("create output dir: %v", err),
		}, nil
	}

	if err := os.WriteFile(filePath, body, 0644); err != nil {
		return &Result{
			ID:      id,
			Title:   title,
			Success: false,
			Error:   fmt.Sprintf("write file: %v", err),
		}, nil
	}

	return &Result{
		ID:       id,
		Title:    title,
		FilePath: filePath,
		Success:  true,
	}, nil
}

// DownloadByDetailPage fetches the detail page then downloads the document.
func (d *Downloader) DownloadByDetailPage(detailLink string, title string) (*Result, error) {
	status, body, err := d.cl.Get(detailLink)
	if err != nil {
		return &Result{
			Title:   title,
			Success: false,
			Error:   fmt.Sprintf("fetch detail page: %v", err),
		}, nil
	}

	if status != 200 {
		return &Result{
			Title:   title,
			Success: false,
			Error:   fmt.Sprintf("detail page returned status %d", status),
		}, nil
	}

	downloadID := extractDownloadIDFromHTML(string(body))
	if downloadID == "" {
		return &Result{
			Title:   title,
			Success: false,
			Error:   "could not extract download ID from detail page",
		}, nil
	}

	fmt.Printf("    Extracted download ID: %s\n", downloadID)
	return d.DownloadByID(downloadID, title)
}

// DownloadBatch downloads multiple documents in sequence.
func (d *Downloader) DownloadBatch(downloads []struct {
	ID    string
	Title string
}) []*Result {
	var results []*Result
	for i, dl := range downloads {
		fmt.Printf("  [%d/%d] Downloading: %s (ID: %s)\n", i+1, len(downloads), dl.Title, dl.ID)
		result, err := d.DownloadByID(dl.ID, dl.Title)
		if err != nil {
			fmt.Printf("    Error: %v\n", err)
			continue
		}
		results = append(results, result)
		if result.Success {
			fmt.Printf("    Saved: %s\n", result.FilePath)
		} else {
			fmt.Printf("    Failed: %s\n", result.Error)
		}
	}
	return results
}

func extractDownloadIDFromHTML(html string) string {
	re := regexp.MustCompile(`<input[^>]*name=["']id["'][^>]*value=["']([a-f0-9]+)["']`)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		return m[1]
	}

	re2 := regexp.MustCompile(`download\s*\(\s*['"]([a-f0-9]+)['"]\s*\)`)
	m2 := re2.FindStringSubmatch(html)
	if len(m2) >= 2 {
		return m2[1]
	}

	re3 := regexp.MustCompile(`data-id\s*=\s*["']([a-f0-9]+)["']`)
	m3 := re3.FindStringSubmatch(html)
	if len(m3) >= 2 {
		return m3[1]
	}

	re4 := regexp.MustCompile(`id=([a-f0-9]{32})`)
	m4 := re4.FindStringSubmatch(html)
	if len(m4) >= 2 {
		return m4[1]
	}

	return ""
}

func detectFileExt(data []byte, title string) string {
	if len(data) > 4 && string(data[:4]) == "%PDF" {
		return ".pdf"
	}
	if len(data) > 4 && data[0] == 0xD0 && data[1] == 0xCF && data[2] == 0x11 && data[3] == 0xE0 {
		return ".doc"
	}
	if len(data) > 4 && data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
		return ".docx"
	}
	if len(data) > 6 && string(data[:6]) == "<!DOCT" {
		return ".html"
	}
	return ".doc"
}

func sanitizeFilename(name string) string {
	invalid := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r"}
	result := name
	for _, ch := range invalid {
		result = strings.ReplaceAll(result, ch, "_")
	}
	result = strings.TrimSpace(result)
	if len(result) > 100 {
		result = result[:100]
	}
	return result
}
