package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ExamResource represents a downloadable exam document (PDF).
type ExamResource struct {
	Year        string `json:"year"`
	EraName     string `json:"era_name"`
	Type        string `json:"type"`      // am2, pm1, pm2
	Kind        string `json:"kind"`      // qs, ans, cmnt
	URL         string `json:"url"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	Downloaded  bool   `json:"downloaded"`
}

// YearConfig defines known IPA exam years.
type YearConfig struct {
	Year    string
	EraCode string
	EraName string
}

// SupportedYears provides standard past exam year definitions.
var SupportedYears = []YearConfig{
	{Year: "2024", EraCode: "2024r06", EraName: "令和6年度 秋期"},
	{Year: "2023", EraCode: "2023r05", EraName: "令和5年度 秋期"},
	{Year: "2022", EraCode: "2022r04", EraName: "令和4年度 秋期"},
	{Year: "2021", EraCode: "2021r03", EraName: "令和3年度 秋期"},
	{Year: "2020", EraCode: "2020r02", EraName: "令和2年度 10月"},
}

// DiscoverExamResources finds all DB exam PDF links for the specified year.
func DiscoverExamResources(targetYear string, examTypeFilter string) ([]ExamResource, error) {
	var results []ExamResource

	for _, yc := range SupportedYears {
		if targetYear != "" && targetYear != "all" && yc.Year != targetYear {
			continue
		}

		pageURL := fmt.Sprintf("https://www.ipa.go.jp/shiken/mondai-kaiotu/%s.html", yc.EraCode)
		links, err := scrapePDFLinksFromIPA(pageURL)
		if err != nil || len(links) == 0 {
			// Fallback to static mapping if dynamic fetch fails
			links = getStaticLinksForYear(yc.Year)
		}

		for _, rawURL := range links {
			res := parseResourceFromURL(rawURL, yc)
			if res == nil {
				continue
			}
			if examTypeFilter != "" && examTypeFilter != "all" && res.Type != examTypeFilter {
				continue
			}
			results = append(results, *res)
		}
	}

	return results, nil
}

func scrapePDFLinksFromIPA(pageURL string) ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, pageURL)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	htmlContent := string(bodyBytes)
	re := regexp.MustCompile(`href="([^"]*_db_[^"]*\.pdf)"`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var links []string
	seen := make(map[string]bool)
	for _, m := range matches {
		link := m[1]
		if !strings.HasPrefix(link, "http") {
			link = "https://www.ipa.go.jp" + link
		}
		if !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}
	return links, nil
}

func parseResourceFromURL(rawURL string, yc YearConfig) *ExamResource {
	parts := strings.Split(rawURL, "/")
	fileName := parts[len(parts)-1]

	// Expected filename pattern: <year>_<db>_<am2|pm1|pm2>_<qs|ans|cmnt>.pdf
	lower := strings.ToLower(fileName)
	if !strings.Contains(lower, "_db_") {
		return nil
	}

	examType := "unknown"
	if strings.Contains(lower, "_am2_") {
		examType = "am2"
	} else if strings.Contains(lower, "_pm1_") {
		examType = "pm1"
	} else if strings.Contains(lower, "_pm2_") {
		examType = "pm2"
	}

	kind := "unknown"
	if strings.HasSuffix(lower, "_qs.pdf") {
		kind = "qs"
	} else if strings.HasSuffix(lower, "_ans.pdf") {
		kind = "ans"
	} else if strings.HasSuffix(lower, "_cmnt.pdf") {
		kind = "cmnt"
	}

	return &ExamResource{
		Year:     yc.Year,
		EraName:  yc.EraName,
		Type:     examType,
		Kind:     kind,
		URL:      rawURL,
		FileName: fileName,
	}
}

func getStaticLinksForYear(year string) []string {
	// Fallback mappings for robust offline / degraded network execution
	switch year {
	case "2024":
		base := "https://www.ipa.go.jp/shiken/mondai-kaiotu/m42obm000000afqx-att/"
		return []string{
			base + "2024r06a_db_am2_qs.pdf",
			base + "2024r06a_db_am2_ans.pdf",
			base + "2024r06a_db_pm1_qs.pdf",
			base + "2024r06a_db_pm1_ans.pdf",
			base + "2024r06a_db_pm1_cmnt.pdf",
			base + "2024r06a_db_pm2_qs.pdf",
			base + "2024r06a_db_pm2_ans.pdf",
			base + "2024r06a_db_pm2_cmnt.pdf",
		}
	case "2023":
		base := "https://www.ipa.go.jp/shiken/mondai-kaiotu/ps6vr70000010d6y-att/"
		return []string{
			base + "2023r05a_db_am2_qs.pdf",
			base + "2023r05a_db_am2_ans.pdf",
			base + "2023r05a_db_pm1_qs.pdf",
			base + "2023r05a_db_pm1_ans.pdf",
			base + "2023r05a_db_pm1_cmnt.pdf",
			base + "2023r05a_db_pm2_qs.pdf",
			base + "2023r05a_db_pm2_ans.pdf",
			base + "2023r05a_db_pm2_cmnt.pdf",
		}
	default:
		return nil
	}
}

// DownloadResource downloads a single ExamResource into destDir/<Year>/<FileName>.
func DownloadResource(destBaseDir string, res ExamResource) (int64, error) {
	targetDir := filepath.Join(destBaseDir, res.Year)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	targetPath := filepath.Join(targetDir, res.FileName)

	// If file already exists and size > 0, check if we can reuse
	if fi, err := os.Stat(targetPath); err == nil && fi.Size() > 0 {
		return fi.Size(), nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(res.URL)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received HTTP status %d", resp.StatusCode)
	}

	tmpFile := targetPath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}

	written, err := io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		_ = os.Remove(tmpFile)
		return 0, fmt.Errorf("failed to write content: %w", err)
	}

	if err := os.Rename(tmpFile, targetPath); err != nil {
		return 0, fmt.Errorf("failed to rename file: %w", err)
	}

	return written, nil
}
