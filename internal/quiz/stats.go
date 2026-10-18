package quiz

import (
	"fmt"
	"io"
	"math"
	"sort"
)

// CategoryWeights defines the standard IPA Exam Category Weights (Morning II).
var CategoryWeights = map[string]float64{
	"正規化理論":            0.28,
	"トランザクション・同時実行制御": 0.24,
	"関係モデル・代数":        0.16,
	"障害回復":             0.12,
	"物理設計・インデックス":     0.12,
	"SQL・整合性制約":        0.08,
}

// CalculatePredictedScore calculates predicted score, status, and confidence based on history.
func CalculatePredictedScore(history []HistoryRecord) ScorePrediction {
	if len(history) == 0 {
		return ScorePrediction{
			Score:      0.0,
			Status:     "未受検",
			Confidence: "なし",
		}
	}

	total := len(history)
	var confidence string
	if total < 10 {
		confidence = "低 (Low / 参考値: 回答数10問未満)"
	} else if total < 20 {
		confidence = "中 (Medium / 暫定推論: 回答数10〜19問)"
	} else {
		confidence = "高 (High / 高精度推論: 回答数20問以上)"
	}

	// Recent answers (up to 30) for trend
	startIndex := 0
	if len(history) > 30 {
		startIndex = len(history) - 30
	}
	recentHistory := history[startIndex:]

	catStats := make(map[string]struct{ total, correct int })
	recentCorrectCount := 0
	for _, h := range recentHistory {
		cat := h.Category
		if cat == "" {
			cat = "その他"
		}
		st := catStats[cat]
		st.total++
		if h.IsCorrect {
			st.correct++
			recentCorrectCount++
		}
		catStats[cat] = st
	}

	overallRecentRate := float64(recentCorrectCount) / float64(len(recentHistory))

	predicted := 0.0
	for cat, weight := range CategoryWeights {
		rate := overallRecentRate
		if st, ok := catStats[cat]; ok && st.total > 0 {
			rate = float64(st.correct) / float64(st.total)
		}
		predicted += rate * weight * 100.0
	}

	roundedScore := math.Round(predicted*10) / 10

	var status string
	if roundedScore >= 80.0 {
		status = "合格安全圏 (Safe Pass)"
	} else if roundedScore >= 65.0 {
		status = "合格圏 (Passing)"
	} else if roundedScore >= 55.0 {
		status = "合格ボーダー (Borderline: あと一歩)"
	} else {
		status = "要基礎強化 (Needs Review: 基準点未達)"
	}

	return ScorePrediction{
		Score:      roundedScore,
		Status:     status,
		Confidence: confidence,
	}
}

// CalculateCategoryStats calculates aggregated statistics per category.
func CalculateCategoryStats(history []HistoryRecord) []CategoryStat {
	catMap := make(map[string]*CategoryStat)

	for _, h := range history {
		cat := h.Category
		if cat == "" {
			cat = "その他"
		}
		st, exists := catMap[cat]
		if !exists {
			st = &CategoryStat{Category: cat}
			catMap[cat] = st
		}
		st.Total++
		if h.IsCorrect {
			st.Correct++
		}
	}

	statsList := make([]CategoryStat, 0, len(catMap))
	for _, st := range catMap {
		if st.Total > 0 {
			st.Rate = (float64(st.Correct) / float64(st.Total)) * 100.0
		}
		if st.Rate < 60.0 {
			st.Status = "要復習"
		} else if st.Rate >= 80.0 {
			st.Status = "合格圏"
		} else {
			st.Status = "標準"
		}
		statsList = append(statsList, *st)
	}

	sort.Slice(statsList, func(i, j int) bool {
		if statsList[i].Rate != statsList[j].Rate {
			return statsList[i].Rate < statsList[j].Rate
		}
		return statsList[i].Total > statsList[j].Total
	})

	return statsList
}

// GetQuestionStats computes statistics for each question in history.
func GetQuestionStats(history []HistoryRecord) map[string]QuestionStat {
	qMap := make(map[string]QuestionStat)
	for _, h := range history {
		if h.QuestionID == "" {
			continue
		}
		st := qMap[h.QuestionID]
		st.QuestionID = h.QuestionID
		st.Total++
		if h.IsCorrect {
			st.Correct++
			st.LatestCorrect = true
		} else {
			st.LatestCorrect = false
		}
		qMap[h.QuestionID] = st
	}
	return qMap
}

// RenderStats outputs the full monitoring report to writer.
func RenderStats(w io.Writer, questions []Question, history []HistoryRecord) {
	if len(history) == 0 {
		fmt.Fprintln(w, "回答履歴はまだありません。クイズを実行すると自動的に記録されます。")
		return
	}

	qMap := make(map[string]Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	totalAnswers := len(history)
	correctAnswers := 0
	for _, h := range history {
		if h.IsCorrect {
			correctAnswers++
		}
	}
	overallRate := (float64(correctAnswers) / float64(totalAnswers)) * 100.0

	pred := CalculatePredictedScore(history)

	fmt.Fprintln(w, "=================================================================")
	fmt.Fprintln(w, " データベーススペシャリスト (DB) 学習履歴・得点能力モニタリング")
	fmt.Fprintln(w, "=================================================================")
	fmt.Fprintf(w, "・累計回答数: %d 問\n", totalAnswers)
	fmt.Fprintf(w, "・累計正解数: %d 問 (通算正解率: %.1f%%)\n", correctAnswers, overallRate)
	fmt.Fprintln(w, "-----------------------------------------------------------------")
	fmt.Fprintf(w, "★ 午前II 予想得点:  %.1f 点 / 100点満点 (基準点: 60点)\n", pred.Score)
	fmt.Fprintf(w, "★ 現在の実力判定:  %s\n", pred.Status)
	fmt.Fprintf(w, "★ 推論の信頼度:    %s\n", pred.Confidence)
	fmt.Fprintln(w, "=================================================================")
	fmt.Fprintln(w)

	// Category stats
	catStats := CalculateCategoryStats(history)
	fmt.Fprintln(w, "【分野別 正解率・習熟度】")
	fmt.Fprintf(w, "%-24s | %-6s | %-6s | %-8s | %s\n", "分野", "回答数", "正解数", "正解率", "判定")
	fmt.Fprintln(w, "-----------------------------------------------------------------")
	for _, cs := range catStats {
		fmt.Fprintf(w, "%-22s | %-8d | %-8d | %6.1f%%  | %s\n", cs.Category, cs.Total, cs.Correct, cs.Rate, cs.Status)
	}
	fmt.Fprintln(w, "-----------------------------------------------------------------")
	fmt.Fprintln(w)

	// Wrong questions
	qStats := GetQuestionStats(history)
	var wrongIDs []string
	for qid, st := range qStats {
		if !st.LatestCorrect {
			wrongIDs = append(wrongIDs, qid)
		}
	}
	sort.Strings(wrongIDs)

	fmt.Fprintf(w, "【直近で不正解の弱点問題 (要復習)】: %d 問\n", len(wrongIDs))
	for _, qid := range wrongIDs {
		st := qStats[qid]
		q, found := qMap[qid]
		cat := "不明"
		qText := ""
		if found {
			cat = q.Category
			runes := []rune(q.Question)
			if len(runes) > 35 {
				qText = string(runes[:35]) + "..."
			} else {
				qText = string(runes)
			}
		}
		fmt.Fprintf(w, "・[%s] (%s) 正解率: %d/%d - %s\n", qid, cat, st.Correct, st.Total, qText)
	}
	fmt.Fprintln(w, "=================================================================")
	fmt.Fprintln(w, "※ '--review-wrong' (-r) オプションで上記弱点問題を集中的に演習できます。")
}
