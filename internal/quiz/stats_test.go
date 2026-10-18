package quiz

import (
	"testing"
)

var testWeights = map[string]float64{
	"正規化理論": 0.5,
	"障害回復": 0.5,
}

func TestCalculatePredictedScore_Empty(t *testing.T) {
	pred := CalculatePredictedScore(nil, testWeights)
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
	pred := CalculatePredictedScore(history, testWeights)
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

func TestShuffleQuestion(t *testing.T) {
	q := Question{
		ID:          "q-test",
		Question:    "Test Question",
		Options:     []string{"Opt A", "Opt B", "Opt C", "Opt D"},
		AnswerIndex: 2, // "Opt C"
	}

	sq := ShuffleQuestion(q)
	if len(sq.ShuffledOptions) != 4 {
		t.Fatalf("expected 4 options, got %d", len(sq.ShuffledOptions))
	}

	// Verify new answer index matches "Opt C"
	if sq.ShuffledOptions[sq.NewAnswerIndex] != "Opt C" {
		t.Errorf("expected new answer index to point to 'Opt C', got %s", sq.ShuffledOptions[sq.NewAnswerIndex])
	}

	// Verify reverse mapping
	origIdx := sq.ResolveOriginalChoice(sq.NewAnswerIndex)
	if origIdx != 2 {
		t.Errorf("expected reverse mapped original index to be 2, got %d", origIdx)
	}
}
