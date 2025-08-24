package tracker

import (
	"errors"
	"fmt"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/models"
	"github.com/AndriyBarskyi/gotrack/internal/storage"
)

// SessionManager handles session-related operations
type SessionManager struct {
	storage storage.Storage
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager(storage storage.Storage) *SessionManager {
	return &SessionManager{
		storage: storage,
	}
}

// Start starts a new session.
func (sm *SessionManager) Start(task string) (*models.Session, error) {
	if task == "" {
		return nil, fmt.Errorf("task name cannot be empty")
	}

	lastSession, err := sm.storage.GetLast()
	if err != nil && !errors.Is(err, models.ErrNoSessions) {
		return nil, fmt.Errorf("error checking existing sessions: %v", err)
	}

	if lastSession != nil && lastSession.EndTime.IsZero() {
		return nil, fmt.Errorf("error starting a new session! Previous task '%v' is not finished", lastSession.Task)
	}

	session := &models.Session{
		Task:      task,
		StartTime: time.Now(),
	}

	if err := sm.storage.Save(session); err != nil {
		return nil, fmt.Errorf("error starting the session: %v", err)
	}

	return session, nil
}

// Finish ends the last session.
func (sm *SessionManager) Finish() (*models.Session, error) {
	sessions, err := sm.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error retrieving sessions: %v", err)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no active session to finish")
	}

	lastSession := sessions[len(sessions)-1]
	if !lastSession.EndTime.IsZero() {
		return nil, fmt.Errorf("error ending the session! Task '%v' is already finished", lastSession.Task)
	}

	lastSession.EndTime = time.Now()

	err = sm.storage.Save(&lastSession)
	if err != nil {
		return nil, fmt.Errorf("error saving finished session: %v", err)
	}

	return &lastSession, nil
}

// GetLast returns the most recent session.
func (sm *SessionManager) GetLast() (*models.Session, error) {
	session, err := sm.storage.GetLast()
	if err != nil {
		return nil, fmt.Errorf("error getting last session: %w", err)
	}
	return session, nil
}

// GetTodaySessions returns all sessions that started today.
func (sm *SessionManager) GetTodaySessions() ([]models.Session, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	sessions, err := sm.storage.GetByDateRange(startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("error getting today's sessions: %w", err)
	}

	return sessions, nil
}

// GetAllSessions returns all sessions.
func (sm *SessionManager) GetAllSessions() ([]models.Session, error) {
	sessions, err := sm.storage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error getting all sessions: %w", err)
	}

	return sessions, nil
}

// GetSessionsForTask returns all sessions for a specific task.
func (sm *SessionManager) GetSessionsForTask(task string) ([]models.Session, error) {
	sessions, err := sm.storage.GetByTask(task)
	if err != nil {
		return nil, fmt.Errorf("error getting sessions for task: %w", err)
	}

	return sessions, nil
}

// FormatSession returns a formatted string representation of a session.
func (sm *SessionManager) FormatSession(ssn models.Session, i int, ssns []models.Session) string {
	endTime := ""
	if !ssn.EndTime.IsZero() {
		endTime = ssn.EndTime.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("Task: %s (%d/%d)\nStart time: %s\nEnd time: %s\n\n",
		ssn.Task,
		i+1,
		len(ssns),
		ssn.StartTime.Format("2006-01-02 15:04:05"),
		endTime,
	)
}
