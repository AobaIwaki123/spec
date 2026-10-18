package quiz

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

// RunOptions configures the interactive quiz session.
type RunOptions struct {
	Count       int
	Category    string
	ReviewWrong bool
	All         bool
}

// RunInteractiveSession executes the interactive CLI quiz with shuffled options.
func RunInteractiveSession(store *Store, in io.Reader, out io.Writer, opts RunOptions) error {
	questions, err := store.LoadQuestions()
	if err != nil {
		return err
	}
	if len(questions) == 0 {
		return fmt.Errorf("no questions found")
	}

	allCategories := getUniqueCategories(questions)

	if opts.ReviewWrong {
		history, err := store.LoadHistory()
		if err != nil {
			return err
		}
		qStats := GetQuestionStats(history)
		wrongIDs := make(map[string]bool)
		for qid, st := range qStats {
			if !st.LatestCorrect {
				wrongIDs[qid] = true
			}
		}

		if len(wrongIDs) == 0 {
			fmt.Fprintln(out, "直近で間違えた問題はありません！素晴らしい成果です。")
			return nil
		}

		var filtered []Question
		for _, q := range questions {
			if wrongIDs[q.ID] {
				filtered = append(filtered, q)
			}
		}
		questions = filtered
		fmt.Fprintf(out, ">> 弱点復習モード: 過去に間違えた %d 問から出題します。\n", len(questions))
	} else if opts.Category != "" {
		var filtered []Question
		target := strings.ToLower(opts.Category)
		for _, q := range questions {
			if strings.Contains(strings.ToLower(q.Category), target) {
				filtered = append(filtered, q)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(out, "No questions found for category: '%s'\n", opts.Category)
			fmt.Fprintf(out, "Available categories: %s\n", strings.Join(allCategories, ", "))
			return nil
		}
		questions = filtered
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	numToAsk := len(questions)
	if !opts.All && !opts.ReviewWrong {
		if opts.Count > 0 && opts.Count < numToAsk {
			numToAsk = opts.Count
		}
	}
	selected := questions[:numToAsk]

	fmt.Fprintln(out, "============================================================")
	fmt.Fprintf(out, " %s 午前II対策 択一クイズ (選択肢常時シャッフル)\n", store.ExamTitle())
	fmt.Fprintf(out, " 出題数: %d問 (回答履歴は自動保存されます)\n", numToAsk)
	fmt.Fprintln(out, "============================================================")
	fmt.Fprintln(out)

	scanner := bufio.NewScanner(in)
	correctCount := 0

	for i, q := range selected {
		// Always randomize options for this question
		sq := ShuffleQuestion(q)

		fmt.Fprintf(out, "[第 %d 問 / 全 %d 問] (分野: %s)\n", i+1, numToAsk, q.Category)
		fmt.Fprintf(out, "ID: %s\n", q.ID)
		fmt.Fprintln(out, q.Question)
		fmt.Fprintln(out)
		for idx, opt := range sq.ShuffledOptions {
			fmt.Fprintf(out, "  %d. %s\n", idx+1, opt)
		}
		fmt.Fprintln(out)

		userChoice := -1
		for userChoice < 0 {
			fmt.Fprint(out, "あなたの解答番号 (1-4, qで中断): ")
			if !scanner.Scan() {
				fmt.Fprintln(out, "\n中断しました。")
				return nil
			}
			input := strings.TrimSpace(scanner.Text())
			if strings.ToLower(input) == "q" {
				fmt.Fprintln(out, "\nクイズを中断しました。")
				return nil
			}
			switch input {
			case "1":
				userChoice = 0
			case "2":
				userChoice = 1
			case "3":
				userChoice = 2
			case "4":
				userChoice = 3
			default:
				fmt.Fprintln(out, "1〜4の番号を入力してください。")
			}
		}

		isCorrect := (userChoice == sq.NewAnswerIndex)
		originalChoice := sq.ResolveOriginalChoice(userChoice)

		// Record to history using original indices for consistent persistent history
		jst := time.FixedZone("JST", 9*60*60)
		_ = store.AppendHistory(HistoryRecord{
			Timestamp:       time.Now().In(jst).Format("2006-01-02T15:04:05.999999-07:00"),
			QuestionID:      q.ID,
			Category:        q.Category,
			UserChoiceIndex: originalChoice,
			CorrectIndex:    q.AnswerIndex,
			IsCorrect:       isCorrect,
			Source:          "cli",
		})

		if isCorrect {
			fmt.Fprintln(out, "\n>> 正解！")
			correctCount++
		} else {
			correctOpt := sq.ShuffledOptions[sq.NewAnswerIndex]
			fmt.Fprintf(out, "\n>> 不正解... (正解は %d. %s)\n", sq.NewAnswerIndex+1, correctOpt)
		}

		fmt.Fprintln(out, "--------------------------------------------------")
		fmt.Fprintln(out, "【解説】")
		fmt.Fprintln(out, q.Explanation)
		fmt.Fprintln(out, "============================================================")
		fmt.Fprintln(out)
	}

	rate := (float64(correctCount) / float64(numToAsk)) * 100.0
	fmt.Fprintln(out, "============================================================")
	fmt.Fprintf(out, " 結果発表: %d / %d 問 正解 (正解率: %.1f%%)\n", correctCount, numToAsk, rate)
	if rate >= 80.0 {
		fmt.Fprintln(out, " 評価: 素晴らしい理解度です！午前II基準点（60%）を大幅にクリアしています。")
	} else if rate >= 60.0 {
		fmt.Fprintln(out, " 評価: 合格ライン（60%）到達です。間違えた分野を重点的に復習しましょう。")
	} else {
		fmt.Fprintln(out, " 評価: 基礎知識の定着が必要です。解説とノートを見直しましょう。")
	}
	fmt.Fprintf(out, " 履歴は %s に記録されました。\n", store.HistoryPath())
	fmt.Fprintf(out, " 'spec %s stats' で分析を確認できます。\n", store.ExamID())
	fmt.Fprintln(out, "============================================================")

	return nil
}

func getUniqueCategories(questions []Question) []string {
	seen := make(map[string]bool)
	var list []string
	for _, q := range questions {
		if !seen[q.Category] {
			seen[q.Category] = true
			list = append(list, q.Category)
		}
	}
	return list
}
