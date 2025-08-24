package tracker_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/AndriyBarskyi/gotrack/internal/models"
	"github.com/AndriyBarskyi/gotrack/internal/tracker"
)

// MockStorage is a mock implementation of the storage.Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(session *models.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockStorage) GetLast() (*models.Session, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockStorage) GetAll() ([]models.Session, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockStorage) GetByDateRange(start, end time.Time) ([]models.Session, error) {
	args := m.Called(start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockStorage) GetByTask(task string) ([]models.Session, error) {
	args := m.Called(task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func TestNewSessionManager(t *testing.T) {
	mockStorage := new(MockStorage)
	sm := tracker.NewSessionManager(mockStorage)
	assert.NotNil(t, sm, "SessionManager should not be nil")
}

func TestSessionManager_Start(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		task        string
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful start",
			setupMock: func(ms *MockStorage) {
				ms.On("GetLast").Return((*models.Session)(nil), models.ErrNoSessions).Once()
				ms.On("Save", mock.AnythingOfType("*models.Session")).Return(nil).Once()
			},
			task:        "test task",
			expectError: false,
		},
		{
			name: "empty task name",
			setupMock: func(ms *MockStorage) {
			},
			task:        "",
			expectError: true,
			errorMsg:    "task name cannot be empty",
		},
		{
			name: "previous session not finished",
			setupMock: func(ms *MockStorage) {
				ms.On("GetLast").Return(&models.Session{
					Task:      "unfinished task",
					StartTime: now.Add(-time.Hour),
				}, nil).Once()
			},
			task:        "new task",
			expectError: true,
			errorMsg:    "error starting a new session! Previous task 'unfinished task' is not finished",
		},
		{
			name: "storage error on save",
			setupMock: func(ms *MockStorage) {
				ms.On("GetLast").Return((*models.Session)(nil), models.ErrNoSessions).Once()
				ms.On("Save", mock.AnythingOfType("*models.Session")).Return(errors.New("save error")).Once()
			},
			task:        "test task",
			expectError: true,
			errorMsg:    "error starting the session: save error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.setupMock(mockStorage)

			sm := tracker.NewSessionManager(mockStorage)
			session, err := sm.Start(tt.task)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.Equal(t, tt.task, session.Task)
				assert.False(t, session.StartTime.IsZero())
				assert.True(t, session.EndTime.IsZero())
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSessionManager_Finish(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful finish",
			setupMock: func(ms *MockStorage) {
				ms.On("GetAll").Return([]models.Session{{
					Task:      "test task",
					StartTime: now.Add(-time.Hour),
				}}, nil).Once()
				ms.On("Save", mock.AnythingOfType("*models.Session")).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "no active session",
			setupMock: func(ms *MockStorage) {
				ms.On("GetAll").Return([]models.Session{}, nil).Once()
			},
			expectError: true,
			errorMsg:    "no active session to finish",
		},
		{
			name: "session already finished",
			setupMock: func(ms *MockStorage) {
				ms.On("GetAll").Return([]models.Session{{
					Task:      "test task",
					StartTime: now.Add(-2 * time.Hour),
					EndTime:   now.Add(-time.Hour),
				}}, nil).Once()
			},
			expectError: true,
			errorMsg:    "error ending the session! Task 'test task' is already finished",
		},
		{
			name: "storage error on save",
			setupMock: func(ms *MockStorage) {
				ms.On("GetAll").Return([]models.Session{{
					Task:      "test task",
					StartTime: now.Add(-time.Hour),
				}}, nil).Once()
				ms.On("Save", mock.AnythingOfType("*models.Session")).Return(errors.New("save error")).Once()
			},
			expectError: true,
			errorMsg:    "error saving finished session: save error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.setupMock(mockStorage)

			sm := tracker.NewSessionManager(mockStorage)
			session, err := sm.Finish()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.False(t, session.EndTime.IsZero())
				assert.True(t, session.EndTime.After(session.StartTime))
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSessionManager_GetLast(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful get last",
			setupMock: func(ms *MockStorage) {
				ms.On("GetLast").Return(&models.Session{
					Task:      "test task",
					StartTime: now.Add(-time.Hour),
					EndTime:   now,
				}, nil).Once()
			},
			expectError: false,
		},
		{
			name: "no sessions",
			setupMock: func(ms *MockStorage) {
				ms.On("GetLast").Return((*models.Session)(nil), models.ErrNoSessions).Once()
			},
			expectError: true,
			errorMsg:    "error getting last session: no sessions found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.setupMock(mockStorage)

			sm := tracker.NewSessionManager(mockStorage)
			session, err := sm.GetLast()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSessionManager_GetAllSessions(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		expectError bool
		errorMsg    string
		expectCount int
	}{
		{
			name: "successful get all sessions",
			setupMock: func(ms *MockStorage) {
				sessions := []models.Session{
					{Task: "task 1", StartTime: time.Now().Add(-2 * time.Hour), EndTime: time.Now().Add(-time.Hour)},
					{Task: "task 2", StartTime: time.Now().Add(-3 * time.Hour), EndTime: time.Now().Add(-2 * time.Hour)},
				}
				ms.On("GetAll").Return(sessions, nil).Once()
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "no sessions",
			setupMock: func(ms *MockStorage) {
				ms.On("GetAll").Return([]models.Session{}, nil).Once()
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "storage error",
			setupMock: func(ms *MockStorage) {
				ms.On("GetAll").Return(([]models.Session)(nil), errors.New("storage error")).Once()
			},
			expectError: true,
			errorMsg:    "error getting all sessions: storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.setupMock(mockStorage)

			sm := tracker.NewSessionManager(mockStorage)
			sessions, err := sm.GetAllSessions()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sessions)
				assert.Len(t, sessions, tt.expectCount)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSessionManager_GetSessionsForTask(t *testing.T) {
	taskName := "test task"
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		task        string
		expectError bool
		errorMsg    string
		expectCount int
	}{
		{
			name: "successful get sessions for task",
			setupMock: func(ms *MockStorage) {
				sessions := []models.Session{
					{Task: taskName, StartTime: time.Now().Add(-2 * time.Hour), EndTime: time.Now().Add(-time.Hour)},
				}
				ms.On("GetByTask", taskName).Return(sessions, nil).Once()
			},
			task:        taskName,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "no sessions for task",
			setupMock: func(ms *MockStorage) {
				ms.On("GetByTask", taskName).Return([]models.Session{}, nil).Once()
			},
			task:        taskName,
			expectError: false,
			expectCount: 0,
		},
		{
			name: "empty task name",
			setupMock: func(ms *MockStorage) {
				ms.On("GetByTask", "").Return(([]models.Session)(nil), errors.New("task name cannot be empty")).Once()
			},
			task:        "",
			expectError: true,
			errorMsg:    "error getting sessions for task: task name cannot be empty",
		},
		{
			name: "storage error",
			setupMock: func(ms *MockStorage) {
				ms.On("GetByTask", taskName).Return(([]models.Session)(nil), errors.New("storage error")).Once()
			},
			task:        taskName,
			expectError: true,
			errorMsg:    "error getting sessions for task: storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.setupMock(mockStorage)

			sm := tracker.NewSessionManager(mockStorage)
			sessions, err := sm.GetSessionsForTask(tt.task)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sessions)
				assert.Len(t, sessions, tt.expectCount)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSessionManager_FormatSession(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-time.Hour)
	endTime := now

	tests := []struct {
		name     string
		session  models.Session
		index    int
		sessions []models.Session
		expected string
	}{
		{
			name: "completed session",
			session: models.Session{
				Task:      "test task",
				StartTime: startTime,
				EndTime:   endTime,
			},
			index: 0,
			sessions: []models.Session{{
				Task:      "test task",
				StartTime: startTime,
				EndTime:   endTime,
			}},
			expected: fmt.Sprintf("Task: test task (1/1)\nStart time: %s\nEnd time: %s\n\n",
				startTime.Format("2006-01-02 15:04:05"),
				endTime.Format("2006-01-02 15:04:05")),
		},
		{
			name: "active session",
			session: models.Session{
				Task:      "active task",
				StartTime: startTime,
			},
			index: 1,
			sessions: []models.Session{
				{Task: "another task", StartTime: startTime.Add(-time.Hour), EndTime: startTime},
				{Task: "active task", StartTime: startTime},
			},
			expected: fmt.Sprintf("Task: active task (2/2)\nStart time: %s\nEnd time: \n\n",
				startTime.Format("2006-01-02 15:04:05")),
		},
		{
			name: "session with custom index",
			session: models.Session{
				Task:      "task with index",
				StartTime: startTime,
				EndTime:   endTime,
			},
			index: 5,
			sessions: []models.Session{
				{Task: "task 1"},
				{Task: "task 2"},
				{Task: "task 3"},
				{Task: "task 4"},
				{Task: "task 5"},
				{Task: "task with index", StartTime: startTime, EndTime: endTime},
			},
			expected: fmt.Sprintf("Task: task with index (6/6)\nStart time: %s\nEnd time: %s\n\n",
				startTime.Format("2006-01-02 15:04:05"),
				endTime.Format("2006-01-02 15:04:05")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := tracker.NewSessionManager(nil)
			result := sm.FormatSession(tt.session, tt.index, tt.sessions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSessionManager_GetTodaySessions(t *testing.T) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tests := []struct {
		name        string
		setupMock   func(*MockStorage)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful get today's sessions",
			setupMock: func(ms *MockStorage) {
				todaySessions := []models.Session{
					{Task: "task 1", StartTime: todayStart.Add(9 * time.Hour), EndTime: todayStart.Add(10 * time.Hour)},
					{Task: "task 2", StartTime: todayStart.Add(11 * time.Hour), EndTime: todayStart.Add(12 * time.Hour)},
				}
				ms.On("GetByDateRange", todayStart, todayStart.Add(24*time.Hour)).Return(todaySessions, nil).Once()
			},
			expectError: false,
		},
		{
			name: "no sessions today",
			setupMock: func(ms *MockStorage) {
				ms.On("GetByDateRange", todayStart, todayStart.Add(24*time.Hour)).Return([]models.Session{}, nil).Once()
			},
			expectError: false,
		},
		{
			name: "storage error",
			setupMock: func(ms *MockStorage) {
				ms.On("GetByDateRange", todayStart, todayStart.Add(24*time.Hour)).
					Return(([]models.Session)(nil), errors.New("storage error")).Once()
			},
			expectError: true,
			errorMsg:    "error getting today's sessions: storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.setupMock(mockStorage)

			sm := tracker.NewSessionManager(mockStorage)
			sessions, err := sm.GetTodaySessions()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sessions)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
