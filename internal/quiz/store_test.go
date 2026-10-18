package quiz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AobaIwaki123/spec/internal/exam"
)

func TestStore_AppendAndLoadHistory(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "test_history.jsonl")

	def, _ := exam.Lookup("db")
	store := &Store{
		examDef:     def,
		historyPath: historyPath,
	}

	record1 := HistoryRecord{
		QuestionID:      "db-test-01",
		Category:        "正規化理論",
		UserChoiceIndex: 1,
		CorrectIndex:    1,
		IsCorrect:       true,
		Source:          "test",
	}

	if err := store.AppendHistory(record1); err != nil {
		t.Fatalf("failed to append history: %v", err)
	}

	records, err := store.LoadHistory()
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].QuestionID != "db-test-01" {
		t.Errorf("expected db-test-01, got %s", records[0].QuestionID)
	}
	if !records[0].IsCorrect {
		t.Errorf("expected is_correct=true")
	}
}

func TestRecordHistoryEntry(t *testing.T) {
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "test_history.jsonl")
	questionsPath := filepath.Join(tempDir, "questions.json")

	qJSON := `[
		{
			"id": "db-test-01",
			"category": "関係モデル・代数",
			"question": "Sample Question?",
			"options": ["A", "B", "C", "D"],
			"answer_index": 2,
			"explanation": "Sample explanation"
		}
	]`
	if err := os.WriteFile(questionsPath, []byte(qJSON), 0644); err != nil {
		t.Fatalf("failed to write test questions: %v", err)
	}

	def, _ := exam.Lookup("db")
	store := &Store{
		examDef:       def,
		questionsPath: questionsPath,
		historyPath:   historyPath,
	}

	opts := RecordOptions{
		QuestionID:      "db-test-01",
		UserChoiceIndex: 2,
		CorrectIndex:    -1,
		Source:          "unit-test",
	}

	rec, err := RecordHistoryEntry(store, opts)
	if err != nil {
		t.Fatalf("failed to record history: %v", err)
	}

	if rec.Category != "関係モデル・代数" {
		t.Errorf("expected category to be auto-filled, got %s", rec.Category)
	}
	if rec.CorrectIndex != 2 {
		t.Errorf("expected correct_index to be 2, got %d", rec.CorrectIndex)
	}
	if !rec.IsCorrect {
		t.Errorf("expected is_correct to be true")
	}
}
