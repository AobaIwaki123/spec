package quiz

import (
	"testing"
)

func TestCalculatePredictedScore_Empty(t *testing.T) {
	pred := CalculatePredictedScore(nil)
	if pred.Score != 0.0 {
		t.Errorf("expected 0.0, got %f", pred.Score)
	}
	if pred.Status != "未受検" {
		t.Errorf("expected 未受検, got %s", pred.Status)
	}
}

func TestCalculatePredictedScore_HighRate(t *testing.T) {
	var history []HistoryRecord
	for i := 0; i < 25; i++ {
		history = append(history, HistoryRecord{
			QuestionID: "db-test",
			Category:   "正規化理論",
			IsCorrect:  true,
		})
	}
	pred := CalculatePredictedScore(history)
	if pred.Score < 95.0 {
		t.Errorf("expected score >= 95.0, got %f", pred.Score)
	}
	if pred.Status != "合格安全圏 (Safe Pass)" {
		t.Errorf("expected 合格安全圏 (Safe Pass), got %s", pred.Status)
	}
	if pred.Confidence != "高 (High / 高精度推論: 回答数20問以上)" {
		t.Errorf("unexpected confidence: %s", pred.Confidence)
	}
}

func TestCalculateCategoryStats(t *testing.T) {
	history := []HistoryRecord{
		{Category: "正規化理論", IsCorrect: true},
		{Category: "正規化理論", IsCorrect: false},
		{Category: "障害回復", IsCorrect: true},
	}
	stats := CalculateCategoryStats(history)
	if len(stats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(stats))
	}
	if stats[0].Category != "正規化理論" {
		t.Errorf("expected first category to be 正規化理論, got %s", stats[0].Category)
	}
	if stats[0].Rate != 50.0 {
		t.Errorf("expected 50%%, got %f", stats[0].Rate)
	}
	if stats[0].Status != "要復習" {
		t.Errorf("expected 要復習, got %s", stats[0].Status)
	}
}

func TestGetQuestionStats(t *testing.T) {
	history := []HistoryRecord{
		{QuestionID: "db-001", IsCorrect: false},
		{QuestionID: "db-001", IsCorrect: true},
		{QuestionID: "db-002", IsCorrect: false},
	}
	qStats := GetQuestionStats(history)
	if !qStats["db-001"].LatestCorrect {
		t.Errorf("expected db-001 latest to be correct")
	}
	if qStats["db-002"].LatestCorrect {
		t.Errorf("expected db-002 latest to be incorrect")
	}
}
