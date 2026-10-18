package quiz

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// RecordOptions contains parameters for appending a history record.
type RecordOptions struct {
	QuestionID      string
	UserChoiceIndex int
	CorrectIndex    int
	Category        string
	IsCorrect       *bool
	Source          string
	RawJSON         string
}

// RecordHistoryEntry processes RecordOptions, fills in defaults using question store, and saves.
func RecordHistoryEntry(store *Store, opts RecordOptions) (*HistoryRecord, error) {
	if strings.TrimSpace(opts.RawJSON) != "" {
		var record HistoryRecord
		if err := json.Unmarshal([]byte(opts.RawJSON), &record); err != nil {
			return nil, fmt.Errorf("failed to parse raw JSON: %w", err)
		}
		if record.QuestionID == "" {
			return nil, fmt.Errorf("question_id is required in JSON payload")
		}
		if err := store.AppendHistory(record); err != nil {
			return nil, err
		}
		return &record, nil
	}

	if opts.QuestionID == "" {
		return nil, fmt.Errorf("question ID is required (--id <question-id>)")
	}

	var matchedQ *Question
	questions, err := store.LoadQuestions()
	if err == nil {
		for _, q := range questions {
			if strings.EqualFold(q.ID, opts.QuestionID) {
				qCopy := q
				matchedQ = &qCopy
				break
			}
		}
	}

	category := opts.Category
	if category == "" && matchedQ != nil {
		category = matchedQ.Category
	}
	if category == "" {
		category = "その他"
	}

	correctIdx := opts.CorrectIndex
	if correctIdx < 0 && matchedQ != nil {
		correctIdx = matchedQ.AnswerIndex
	}

	isCorrect := false
	if opts.IsCorrect != nil {
		isCorrect = *opts.IsCorrect
	} else if opts.UserChoiceIndex >= 0 && correctIdx >= 0 {
		isCorrect = (opts.UserChoiceIndex == correctIdx)
	}

	source := opts.Source
	if source == "" {
		source = "chat"
	}

	jst := time.FixedZone("JST", 9*60*60)
	nowISO := time.Now().In(jst).Format("2006-01-02T15:04:05.999999-07:00")

	record := HistoryRecord{
		Timestamp:       nowISO,
		QuestionID:      opts.QuestionID,
		Category:        category,
		UserChoiceIndex: opts.UserChoiceIndex,
		CorrectIndex:    correctIdx,
		IsCorrect:       isCorrect,
		Source:          source,
	}

	if err := store.AppendHistory(record); err != nil {
		return nil, fmt.Errorf("failed to append history: %w", err)
	}

	return &record, nil
}
