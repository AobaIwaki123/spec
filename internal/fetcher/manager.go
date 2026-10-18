package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TaskStatus represents the progress and outcome of a fetch operation.
type TaskStatus struct {
	Status         string         `json:"status"` // idle, running, completed, failed
	StartedAt      string         `json:"started_at,omitempty"`
	FinishedAt     string         `json:"finished_at,omitempty"`
	ExamID         string         `json:"exam_id"`
	TargetYear     string         `json:"target_year"`
	ExamType       string         `json:"exam_type"`
	TotalItems     int            `json:"total_items"`
	CompletedItems int            `json:"completed_items"`
	TotalBytes     int64          `json:"total_bytes"`
	CurrentItem    string         `json:"current_item,omitempty"`
	Errors         []string       `json:"errors,omitempty"`
	Items          []ExamResource `json:"items,omitempty"`
}

// Manager coordinates fetching and tracking.
type Manager struct {
	baseDataDir string
	statusPath  string
}

// NewManager creates a Manager targeting the specified data directory.
func NewManager(baseDataDir string) *Manager {
	return &Manager{
		baseDataDir: baseDataDir,
		statusPath:  filepath.Join(baseDataDir, "fetch_status.json"),
	}
}

// StatusPath returns the task status JSON location.
func (m *Manager) StatusPath() string {
	return m.statusPath
}

// LoadStatus reads the current status from disk.
func (m *Manager) LoadStatus() (*TaskStatus, error) {
	data, err := os.ReadFile(m.statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskStatus{Status: "idle"}, nil
		}
		return nil, err
	}
	var st TaskStatus
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// SaveStatus writes the given status to disk atomically.
func (m *Manager) SaveStatus(st *TaskStatus) error {
	dir := filepath.Dir(m.statusPath)
	_ = os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmpFile := m.statusPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, m.statusPath)
}

// Run executes the download synchronously.
func (m *Manager) Run(examCode, targetYear, examType string, progressFn func(cur, tot int, item string)) (*TaskStatus, error) {
	jst := time.FixedZone("JST", 9*60*60)
	now := time.Now().In(jst).Format(time.RFC3339)

	resources, err := DiscoverExamResources(examCode, targetYear, examType)
	if err != nil {
		st := &TaskStatus{
			Status:     "failed",
			StartedAt:  now,
			FinishedAt: time.Now().In(jst).Format(time.RFC3339),
			ExamID:     examCode,
			TargetYear: targetYear,
			ExamType:   examType,
			Errors:     []string{fmt.Sprintf("discovery failed: %v", err)},
		}
		_ = m.SaveStatus(st)
		return st, err
	}

	st := &TaskStatus{
		Status:         "running",
		StartedAt:      now,
		ExamID:         examCode,
		TargetYear:     targetYear,
		ExamType:       examType,
		TotalItems:     len(resources),
		CompletedItems: 0,
		Items:          resources,
	}
	_ = m.SaveStatus(st)

	for i, res := range resources {
		st.CurrentItem = fmt.Sprintf("[%s %s %s] %s", res.ExamID, res.Year, res.Type, res.FileName)
		_ = m.SaveStatus(st)

		if progressFn != nil {
			progressFn(i+1, len(resources), st.CurrentItem)
		}

		bytes, err := DownloadResource(m.baseDataDir, res)
		if err != nil {
			st.Errors = append(st.Errors, fmt.Sprintf("%s: %v", res.FileName, err))
		} else {
			st.CompletedItems++
			st.TotalBytes += bytes
			res.Size = bytes
			res.Downloaded = true
			st.Items[i] = res
		}
		_ = m.SaveStatus(st)
	}

	st.FinishedAt = time.Now().In(jst).Format(time.RFC3339)
	st.CurrentItem = ""
	if len(st.Errors) > 0 && st.CompletedItems == 0 {
		st.Status = "failed"
	} else {
		st.Status = "completed"
	}
	_ = m.SaveStatus(st)

	return st, nil
}

// StartBackground launches the current binary as a background daemon process.
func (m *Manager) StartBackground(examCode, targetYear, examType string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("unable to determine executable path: %w", err)
	}

	args := []string{examCode, "fetch", "--internal-worker", "--year", targetYear, "--type", examType}

	cmd := exec.Command(exe, args...)
	cmd.Dir = m.baseDataDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start background worker: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	return cmd.Process.Pid, nil
}

// RenderStatus displays a formatted summary of the latest fetch task.
func (m *Manager) RenderStatus(w io.Writer) {
	st, err := m.LoadStatus()
	if err != nil {
		fmt.Fprintf(w, "Error loading status: %v\n", err)
		return
	}

	examCode := st.ExamID
	if examCode == "" {
		examCode = "db"
	}
	fmt.Fprintln(w, "=================================================================")
	fmt.Fprintf(w, " [%s] 過去問PDF非同期取得タスク ステータス\n", strings.ToUpper(examCode))
	fmt.Fprintln(w, "=================================================================")
	fmt.Fprintf(w, "・実行状態: %s\n", st.Status)
	if st.StartedAt != "" {
		fmt.Fprintf(w, "・開始日時: %s\n", st.StartedAt)
	}
	if st.FinishedAt != "" {
		fmt.Fprintf(w, "・完了日時: %s\n", st.FinishedAt)
	}
	fmt.Fprintf(w, "・進捗状況: %d / %d 件完了\n", st.CompletedItems, st.TotalItems)
	if st.TotalBytes > 0 {
		mb := float64(st.TotalBytes) / (1024 * 1024)
		fmt.Fprintf(w, "・総取得容量: %.2f MB\n", mb)
	}
	if st.CurrentItem != "" {
		fmt.Fprintf(w, "・現在処理中: %s\n", st.CurrentItem)
	}
	fmt.Fprintf(w, "・保存先:   %s/%s/exams\n", m.baseDataDir, st.ExamID)

	if len(st.Errors) > 0 {
		fmt.Fprintln(w, "-----------------------------------------------------------------")
		fmt.Fprintf(w, "【エラー件数: %d 件】\n", len(st.Errors))
		for _, e := range st.Errors {
			fmt.Fprintf(w, "  - %s\n", e)
		}
	}

	if len(st.Items) > 0 {
		fmt.Fprintln(w, "-----------------------------------------------------------------")
		fmt.Fprintln(w, "【取得済みファイル一覧】")
		for _, item := range st.Items {
			statusMark := "○"
			if !item.Downloaded {
				statusMark = "×"
			}
			mb := float64(item.Size) / (1024 * 1024)
			fmt.Fprintf(w, " %s [%s %s] %-4s %-4s (%.2f MB) -> %s\n", statusMark, item.ExamID, item.Year, item.Type, item.Kind, mb, item.FileName)
		}
	}
	fmt.Fprintln(w, "=================================================================")
}
