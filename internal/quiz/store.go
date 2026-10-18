package quiz

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AobaIwaki123/spec/internal/exam"
)

// Store manages loading questions and saving/reading quiz history for a specific exam.
type Store struct {
	examDef       exam.Definition
	baseDataDir   string
	examDataDir   string
	questionsPath string
	historyPath   string
}

// NewStore initializes a Store with automatic data directory resolution for the given exam key.
func NewStore(examKey string) *Store {
	def, ok := exam.Lookup(examKey)
	if !ok {
		def, _ = exam.Lookup(string(exam.DefaultExamID))
	}

	baseDataDir := resolveBaseDataDir()
	examDataDir := filepath.Join(baseDataDir, string(def.ID))
	_ = os.MkdirAll(examDataDir, 0755)

	questionsPath := os.Getenv("SPEC_QUESTIONS_PATH")
	if questionsPath == "" {
		candidate := filepath.Join(examDataDir, "questions.json")
		if _, err := os.Stat(candidate); err == nil {
			questionsPath = candidate
		}
	}
	if questionsPath == "" {
		// Fallbacks
		candidates := []string{
			filepath.Join(baseDataDir, "questions.json"),
			fmt.Sprintf("/Users/aobaiwaki/spec/data/%s/questions.json", def.ID),
			"/Users/aobaiwaki/spec/data/questions.json",
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

	historyPath := os.Getenv("SPEC_HISTORY_PATH")
	if historyPath == "" {
		historyPath = filepath.Join(examDataDir, "quiz_history.jsonl")
	}

	return &Store{
		examDef:       def,
		baseDataDir:   baseDataDir,
		examDataDir:   examDataDir,
		questionsPath: questionsPath,
		historyPath:   historyPath,
	}
}

// ExamID returns the active exam identifier (e.g. "db").
func (s *Store) ExamID() string {
	return string(s.examDef.ID)
}

// ExamTitle returns the friendly name of the active exam.
func (s *Store) ExamTitle() string {
	return s.examDef.Name
}

// ExamDef returns the full definition of the active exam.
func (s *Store) ExamDef() exam.Definition {
	return s.examDef
}

// BaseDataDir returns the root data directory.
func (s *Store) BaseDataDir() string {
	return s.baseDataDir
}

// ExamDataDir returns the exam-specific data directory.
func (s *Store) ExamDataDir() string {
	return s.examDataDir
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
		return nil, fmt.Errorf("questions.json file not found for %s (set SPEC_QUESTIONS_PATH)", s.examDef.ID)
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
			// Seed fallback if base history exists
			baseHist := filepath.Join(s.baseDataDir, "quiz_history.jsonl")
			if s.historyPath != baseHist {
				if _, err := os.Stat(baseHist); err == nil {
					_ = s.AppendHistoryFromFile(baseHist)
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

func resolveBaseDataDir() string {
	if env := os.Getenv("SPEC_DATA_DIR"); env != "" {
		return env
	}

	repoRoot := findRepoRoot()
	if repoRoot != "" {
		candidate := filepath.Join(repoRoot, "data")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	defaultRepoData := "/Users/aobaiwaki/spec/data"
	if _, err := os.Stat(defaultRepoData); err == nil {
		return defaultRepoData
	}

	home, err := os.UserHomeDir()
	if err == nil {
		xdgData := filepath.Join(home, ".local", "share", "spec")
		_ = os.MkdirAll(xdgData, 0755)
		return xdgData
	}

	return "./data"
}

func findRepoRoot() string {
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for {
			if isSpecRoot(dir) {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			dir := filepath.Dir(resolved)
			for {
				if isSpecRoot(dir) {
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

func isSpecRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "cmd", "spec")); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return true
		}
	}
	return false
}
