package fetcher

import (
	"path/filepath"
	"testing"
)

func TestParseResourceFromURL(t *testing.T) {
	yc := YearConfig{
		Year:    "2024",
		EraCode: "2024r06",
		EraName: "令和6年度 秋期",
	}

	url := "https://www.ipa.go.jp/shiken/mondai-kaiotu/m42obm000000afqx-att/2024r06a_db_am2_qs.pdf"
	res := parseResourceFromURL("db", url, yc)
	if res == nil {
		t.Fatalf("expected non-nil resource")
	}

	if res.ExamID != "db" {
		t.Errorf("expected db, got %s", res.ExamID)
	}
	if res.Year != "2024" {
		t.Errorf("expected 2024, got %s", res.Year)
	}
	if res.Type != "am2" {
		t.Errorf("expected am2, got %s", res.Type)
	}
	if res.Kind != "qs" {
		t.Errorf("expected qs, got %s", res.Kind)
	}
	if res.FileName != "2024r06a_db_am2_qs.pdf" {
		t.Errorf("unexpected filename: %s", res.FileName)
	}
}

func TestManager_SaveAndLoadStatus(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManager(tempDir)

	st := &TaskStatus{
		Status:         "completed",
		ExamID:         "db",
		TargetYear:     "2024",
		TotalItems:     5,
		CompletedItems: 5,
	}

	if err := mgr.SaveStatus(st); err != nil {
		t.Fatalf("failed to save status: %v", err)
	}

	loaded, err := mgr.LoadStatus()
	if err != nil {
		t.Fatalf("failed to load status: %v", err)
	}

	if loaded.Status != "completed" {
		t.Errorf("expected status completed, got %s", loaded.Status)
	}
	if loaded.CompletedItems != 5 {
		t.Errorf("expected 5 completed, got %d", loaded.CompletedItems)
	}
	if mgr.StatusPath() != filepath.Join(tempDir, "fetch_status.json") {
		t.Errorf("unexpected status path: %s", mgr.StatusPath())
	}
}
