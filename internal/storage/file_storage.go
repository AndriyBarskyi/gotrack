package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/models"
)

// Storage defines the interface for session storage operations
type Storage interface {
	Save(session *models.Session) error
	GetLast() (*models.Session, error)
	GetAll() ([]models.Session, error)
	GetByDateRange(start, end time.Time) ([]models.Session, error)
	GetByTask(task string) ([]models.Session, error)
}

// FileStorage implements the Storage interface using a JSONL file.
type FileStorage struct {
	filePath string
}

// NewFileStorage creates a new FileStorage instance.
func NewFileStorage(filePath string) (*FileStorage, error) {
	if filePath == "" {
		return nil, errors.New("file path cannot be empty")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create/open storage file: %w", err)
	}
	file.Close()

	return &FileStorage{
		filePath: filePath,
	}, nil
}

// Save appends a session to the storage file.
func (s *FileStorage) Save(session *models.Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	if session.Task == "" {
		return errors.New("task name cannot be empty")
	}
	if session.StartTime.IsZero() {
		return errors.New("start time cannot be zero")
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open storage file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to storage file: %w", err)
	}

	return f.Sync()
}

// GetLast returns the most recent session from the storage.
func (s *FileStorage) GetLast() (*models.Session, error) {
	sessions, err := s.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	if len(sessions) == 0 {
		return nil, models.ErrNoSessions
	}

	return &sessions[len(sessions)-1], nil
}

// GetAll returns all sessions from the storage.
func (s *FileStorage) GetAll() ([]models.Session, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Session{}, nil
		}
		return nil, fmt.Errorf("failed to open storage file: %w", err)
	}
	defer file.Close()

	var sessions []models.Session
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var session models.Session
		if err := json.Unmarshal(scanner.Bytes(), &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading storage file: %w", err)
	}

	return sessions, nil
}

// GetByDateRange returns sessions within the specified date range (inclusive).
func (s *FileStorage) GetByDateRange(start, end time.Time) ([]models.Session, error) {
	sessions, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	var result []models.Session
	for _, s := range sessions {
		if (s.StartTime.After(start) || s.StartTime.Equal(start)) &&
			(s.StartTime.Before(end) || s.StartTime.Equal(end)) {
			result = append(result, s)
		}
	}

	return result, nil
}

// GetByTask returns all sessions for the specified task.
func (s *FileStorage) GetByTask(task string) ([]models.Session, error) {
	sessions, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	var result []models.Session
	for _, s := range sessions {
		if s.Task == task {
			result = append(result, s)
		}
	}

	return result, nil
}
