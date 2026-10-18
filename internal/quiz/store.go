package quiz

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Store manages loading questions and saving/reading quiz history.
type Store struct {
	dataDir       string
	questionsPath string
	historyPath   string
}

// NewStore initializes a Store with automatic data directory resolution.
func NewStore() *Store {
	dataDir := resolveDataDir()

	questionsPath := os.Getenv("DB_QUIZ_QUESTIONS_PATH")
	if questionsPath == "" {
		candidate := filepath.Join(dataDir, "questions.json")
		if _, err := os.Stat(candidate); err == nil {
			questionsPath = candidate
		}
	}
	if questionsPath == "" {
		// Fallbacks
		candidates := []string{
			"/Users/aobaiwaki/db-quiz/data/questions.json",
			"/Users/aobaiwaki/ai-workspace/.agents/skills/db-specialist-quiz/references/questions.json",
			"data/questions.json",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				questionsPath = c
				break
			}
		}
	}

	historyPath := os.Getenv("DB_QUIZ_HISTORY_PATH")
	if historyPath == "" {
		historyPath = filepath.Join(dataDir, "quiz_history.jsonl")
	}

	return &Store{
		dataDir:       dataDir,
		questionsPath: questionsPath,
		historyPath:   historyPath,
	}
}

// DataDir returns the base data directory.
func (s *Store) DataDir() string {
	return s.dataDir
}

// QuestionsPath returns the resolved questions.json path.
func (s *Store) QuestionsPath() string {
	return s.questionsPath
}

// HistoryPath returns the resolved quiz_history.jsonl path.
func (s *Store) HistoryPath() string {
	return s.historyPath
}

// LoadQuestions loads questions from the resolved questionsPath.
func (s *Store) LoadQuestions() ([]Question, error) {
	if s.questionsPath == "" {
		return nil, fmt.Errorf("questions.json file not found (set DB_QUIZ_QUESTIONS_PATH)")
	}
	data, err := os.ReadFile(s.questionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read questions file from %s: %w", s.questionsPath, err)
	}

	var questions []Question
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse questions JSON: %w", err)
	}
	return questions, nil
}

// LoadHistory loads all history records from the quiz_history.jsonl file.
func (s *Store) LoadHistory() ([]HistoryRecord, error) {
	file, err := os.Open(s.historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If not found in primary path, check fallback in ai-workspace
			fallbackPath := "/Users/aobaiwaki/ai-workspace/space/database-specialist/data/quiz_history.jsonl"
			if s.historyPath != fallbackPath {
				if fbData, err := os.ReadFile(fallbackPath); err == nil && len(fbData) > 0 {
					// Seed into primary path
					_ = s.AppendHistoryFromFile(fallbackPath)
					return s.LoadHistory()
				}
			}
			return []HistoryRecord{}, nil
		}
		return nil, fmt.Errorf("failed to open history file %s: %w", s.historyPath, err)
	}
	defer file.Close()

	var records []HistoryRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var r HistoryRecord
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			continue
		}
		records = append(records, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading history file: %w", err)
	}
	return records, nil
}

// AppendHistory appends a new record to the quiz_history.jsonl file.
func (s *Store) AppendHistory(record HistoryRecord) error {
	dir := filepath.Dir(s.historyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for history file %s: %w", dir, err)
	}

	if record.Timestamp == "" {
		jst := time.FixedZone("JST", 9*60*60)
		record.Timestamp = time.Now().In(jst).Format("2006-01-02T15:04:05.999999-07:00")
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal history record: %w", err)
	}

	f, err := os.OpenFile(s.historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file for appending: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write history record: %w", err)
	}
	return nil
}

// AppendHistoryFromFile copies records from another history file if target is empty.
func (s *Store) AppendHistoryFromFile(srcPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var r HistoryRecord
		if err := json.Unmarshal([]byte(line), &r); err == nil {
			_ = s.AppendHistory(r)
		}
	}
	return nil
}

// resolveDataDir finds the data directory across environment, repo root, and XDG paths.
func resolveDataDir() string {
	if env := os.Getenv("DB_QUIZ_DATA_DIR"); env != "" {
		return env
	}

	// 1. Check repo root of db-quiz (executable path or current dir)
	repoRoot := findRepoRoot()
	if repoRoot != "" {
		candidate := filepath.Join(repoRoot, "data")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// 2. Default repo location
	defaultRepoData := "/Users/aobaiwaki/db-quiz/data"
	if _, err := os.Stat(defaultRepoData); err == nil {
		return defaultRepoData
	}

	// 3. XDG standard data directory
	home, err := os.UserHomeDir()
	if err == nil {
		xdgData := filepath.Join(home, ".local", "share", "db-quiz")
		_ = os.MkdirAll(xdgData, 0755)
		return xdgData
	}

	return "./data"
}

func findRepoRoot() string {
	// From current working directory upwards
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for {
			if isDBQuizRoot(dir) {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// From executable path
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			dir := filepath.Dir(resolved)
			for {
				if isDBQuizRoot(dir) {
					return dir
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
		}
	}

	return ""
}

func isDBQuizRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "cmd", "db-quiz")); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return true
		}
	}
	return false
}
