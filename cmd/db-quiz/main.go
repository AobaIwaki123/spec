package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/AobaIwaki123/db-quiz/internal/fetcher"
	"github.com/AobaIwaki123/db-quiz/internal/quiz"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	store := quiz.NewStore()

	if len(args) > 0 {
		subcmd := args[0]
		switch subcmd {
		case "stats":
			return handleStats(store)
		case "categories", "list-categories":
			return handleListCategories(store)
		case "review":
			return handleReview(store)
		case "record":
			return handleRecord(store, args[1:])
		case "question", "questions":
			return handleQuestion(store, args[1:])
		case "fetch", "download":
			return handleFetch(store, args[1:])
		case "help", "-h", "--help":
			printUsage()
			return nil
		}
	}

	// Handle root flags
	rootFlags := flag.NewFlagSet("db-quiz", flag.ContinueOnError)
	var (
		count          = rootFlags.Int("n", 5, "Number of questions to ask (default: 5)")
		countLong      = rootFlags.Int("count", 5, "Number of questions to ask (default: 5)")
		category       = rootFlags.String("c", "", "Filter by specific category")
		categoryLong   = rootFlags.String("category", "", "Filter by specific category")
		reviewWrong    = rootFlags.Bool("r", false, "Review questions answered incorrectly last time")
		reviewWrongL   = rootFlags.Bool("review-wrong", false, "Review questions answered incorrectly last time")
		all            = rootFlags.Bool("all", false, "Ask all questions in random order")
		stats          = rootFlags.Bool("stats", false, "Show learning statistics and weak areas")
		listCategories = rootFlags.Bool("list-categories", false, "List all available categories")
		recordJSON     = rootFlags.String("record", "", "Record a quiz attempt (JSON string or pass via 'record' subcommand)")
	)

	if err := rootFlags.Parse(args); err != nil {
		return err
	}

	effectiveCount := *count
	if isFlagPassed(rootFlags, "count") {
		effectiveCount = *countLong
	}

	effectiveCategory := *category
	if isFlagPassed(rootFlags, "category") {
		effectiveCategory = *categoryLong
	}

	effectiveReview := *reviewWrong || *reviewWrongL

	if *stats {
		return handleStats(store)
	}

	if *listCategories {
		return handleListCategories(store)
	}

	if *recordJSON != "" {
		return handleRecord(store, []string{"--json", *recordJSON})
	}

	return quiz.RunInteractiveSession(store, os.Stdin, os.Stdout, quiz.RunOptions{
		Count:       effectiveCount,
		Category:    effectiveCategory,
		ReviewWrong: effectiveReview,
		All:         *all,
	})
}

func isFlagPassed(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func handleStats(store *quiz.Store) error {
	questions, err := store.LoadQuestions()
	if err != nil {
		return err
	}
	history, err := store.LoadHistory()
	if err != nil {
		return err
	}
	quiz.RenderStats(os.Stdout, questions, history)
	return nil
}

func handleListCategories(store *quiz.Store) error {
	questions, err := store.LoadQuestions()
	if err != nil {
		return err
	}
	catCounts := make(map[string]int)
	for _, q := range questions {
		catCounts[q.Category]++
	}

	var cats []string
	for c := range catCounts {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	fmt.Println("Available Categories:")
	for _, c := range cats {
		fmt.Printf("  - %s (%d questions)\n", c, catCounts[c])
	}
	return nil
}

func handleReview(store *quiz.Store) error {
	return quiz.RunInteractiveSession(store, os.Stdin, os.Stdout, quiz.RunOptions{
		ReviewWrong: true,
	})
}

func handleRecord(store *quiz.Store, args []string) error {
	fs := flag.NewFlagSet("db-quiz record", flag.ContinueOnError)
	var (
		id        = fs.String("id", "", "Question ID (e.g. db-001)")
		choice    = fs.Int("choice", -1, "User choice index (0-3)")
		correct   = fs.Int("correct", -1, "Correct answer index (0-3, auto-filled if omitted)")
		category  = fs.String("category", "", "Category name (auto-filled if omitted)")
		isCorrect = fs.String("is-correct", "", "Whether answer was correct ('true' or 'false')")
		source    = fs.String("source", "chat", "Source identifier (e.g. 'chat', 'cli')")
		rawJSON   = fs.String("json", "", "Raw JSON string representing the history record")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	var isCorrectPtr *bool
	if *isCorrect != "" {
		val, err := strconv.ParseBool(*isCorrect)
		if err != nil {
			return fmt.Errorf("invalid --is-correct value: %w", err)
		}
		isCorrectPtr = &val
	}

	opts := quiz.RecordOptions{
		QuestionID:      *id,
		UserChoiceIndex: *choice,
		CorrectIndex:    *correct,
		Category:        *category,
		IsCorrect:       isCorrectPtr,
		Source:          *source,
		RawJSON:         *rawJSON,
	}

	rec, err := quiz.RecordHistoryEntry(store, opts)
	if err != nil {
		return err
	}

	out, _ := json.Marshal(rec)
	fmt.Printf("Recorded: %s\n", string(out))
	return nil
}

func handleQuestion(store *quiz.Store, args []string) error {
	fs := flag.NewFlagSet("db-quiz question", flag.ContinueOnError)
	var (
		id          = fs.String("id", "", "Filter by specific Question ID")
		category    = fs.String("category", "", "Filter by category")
		reviewWrong = fs.Bool("review-wrong", false, "Filter to recently incorrect questions")
		asJSON      = fs.Bool("json", false, "Output results as JSON")
		count       = fs.Int("count", 1, "Number of questions to output (default: 1)")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	questions, err := store.LoadQuestions()
	if err != nil {
		return err
	}

	var filtered []quiz.Question
	if *id != "" {
		for _, q := range questions {
			if strings.EqualFold(q.ID, *id) {
				filtered = append(filtered, q)
				break
			}
		}
	} else if *reviewWrong {
		history, _ := store.LoadHistory()
		qStats := quiz.GetQuestionStats(history)
		for _, q := range questions {
			if st, ok := qStats[q.ID]; ok && !st.LatestCorrect {
				filtered = append(filtered, q)
			}
		}
	} else if *category != "" {
		target := strings.ToLower(*category)
		for _, q := range questions {
			if strings.Contains(strings.ToLower(q.Category), target) {
				filtered = append(filtered, q)
			}
		}
	} else {
		filtered = questions
	}

	if len(filtered) == 0 {
		return fmt.Errorf("no matching questions found")
	}

	limit := *count
	if limit > len(filtered) {
		limit = len(filtered)
	}
	selected := filtered[:limit]

	if *asJSON {
		var out []byte
		if limit == 1 {
			out, _ = json.MarshalIndent(selected[0], "", "  ")
		} else {
			out, _ = json.MarshalIndent(selected, "", "  ")
		}
		fmt.Println(string(out))
		return nil
	}

	for _, q := range selected {
		fmt.Printf("[%s] (%s)\n%s\n", q.ID, q.Category, q.Question)
		for idx, opt := range q.Options {
			fmt.Printf("  %d. %s\n", idx+1, opt)
		}
		fmt.Println()
	}

	return nil
}

func handleFetch(store *quiz.Store, args []string) error {
	fs := flag.NewFlagSet("db-quiz fetch", flag.ContinueOnError)
	var (
		year           = fs.String("year", "all", "Target exam year (e.g. 2024, 2023, or 'all')")
		allFlag        = fs.Bool("all", false, "Fetch all supported years (same as --year all)")
		examType       = fs.String("type", "all", "Exam type filter (am2, pm1, pm2, or 'all')")
		background     = fs.Bool("background", false, "Run download as a background daemon process")
		backgroundD    = fs.Bool("d", false, "Run download as a background daemon process")
		showStatus     = fs.Bool("status", false, "Show current fetch progress and status")
		internalWorker = fs.Bool("internal-worker", false, "Internal background worker flag")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	targetYear := *year
	if *allFlag {
		targetYear = "all"
	}

	mgr := fetcher.NewManager(store.DataDir())

	if *showStatus {
		mgr.RenderStatus(os.Stdout)
		return nil
	}

	if *internalWorker {
		_, err := mgr.Run(targetYear, *examType, nil)
		return err
	}

	runBg := *background || *backgroundD
	if runBg {
		pid, err := mgr.StartBackground(targetYear, *examType)
		if err != nil {
			return err
		}
		fmt.Printf("過去問PDF取得タスクをバックグラウンドで開始しました (PID: %d)。\n", pid)
		fmt.Printf("進行状況の確認: db-quiz fetch --status\n")
		return nil
	}

	fmt.Printf("IPA公式アーカイブから過去問PDFを取得しています (対象年度: %s, 区分: %s)...\n", targetYear, *examType)
	_, err := mgr.Run(targetYear, *examType, func(cur, tot int, item string) {
		fmt.Printf("[%d/%d] 取得中: %s\n", cur, tot, item)
	})
	if err != nil {
		return err
	}

	fmt.Println("\n取得が完了しました。")
	mgr.RenderStatus(os.Stdout)
	return nil
}

func printUsage() {
	fmt.Println(`db-quiz - Database Specialist (IPA DB) Morning II Quiz & Monitoring Tool

Usage:
  db-quiz [options]
  db-quiz <subcommand> [options]

Subcommands:
  stats               Display learning progress, predicted score, and weak topics
  categories          List all available question categories
  review              Review questions answered incorrectly last time
  record [flags]      Record a single question attempt (ideal for chat/agent integration)
  question [flags]    Fetch questions (supports JSON output for agent integration)
  fetch [flags]       Fetch and archive official IPA past exam PDFs (supports background execution)

Options:
  -n, --count <int>   Number of questions to ask (default: 5)
  -c, --category <str> Filter questions by category keyword
  -r, --review-wrong  Review questions answered incorrectly last time
  --all               Ask all available questions
  --stats             Display statistics (same as 'db-quiz stats')
  --list-categories   List categories (same as 'db-quiz categories')
  --record <json>     Record attempt via JSON payload

Fetch Subcommand Flags:
  --year <year>       Target year (e.g. 2024, 2023, 2022, or 'all', default: all)
  --type <type>       Filter exam type (am2, pm1, pm2, or 'all', default: all)
  -d, --background    Run download asynchronously as a background daemon
  --status            Display current download status and archived files`)
}
