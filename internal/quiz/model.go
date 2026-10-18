package quiz

// Question represents a multiple-choice exam question.
type Question struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
}

// HistoryRecord represents a single question attempt in quiz_history.jsonl.
type HistoryRecord struct {
	Timestamp       string `json:"timestamp"`
	QuestionID      string `json:"question_id"`
	Category        string `json:"category"`
	UserChoiceIndex int    `json:"user_choice_index"`
	CorrectIndex    int    `json:"correct_index"`
	IsCorrect       bool   `json:"is_correct"`
	Source          string `json:"source"`
}

// ScorePrediction holds the calculated prediction results.
type ScorePrediction struct {
	Score      float64 `json:"score"`
	Status     string  `json:"status"`
	Confidence string  `json:"confidence"`
}

// CategoryStat holds calculated statistics for a single category.
type CategoryStat struct {
	Category string  `json:"category"`
	Total    int     `json:"total"`
	Correct  int     `json:"correct"`
	Rate     float64 `json:"rate"`
	Status   string  `json:"status"`
}

// QuestionStat holds summary statistics for a specific question ID.
type QuestionStat struct {
	QuestionID    string `json:"question_id"`
	Total         int    `json:"total"`
	Correct       int    `json:"correct"`
	LatestCorrect bool   `json:"latest_correct"`
}
