package storage_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/models"
	"github.com/AndriyBarskyi/gotrack/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFile(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sessions.jsonl")
	return filePath, func() {
		os.RemoveAll(tempDir)
	}
}

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		setup       func() (string, func())
		expectError bool
	}{
		{
			name:        "valid file path",
			filePath:    filepath.Join(t.TempDir(), "sessions.jsonl"),
			setup:       nil,
			expectError: false,
		},
		{
			name:        "empty file path",
			filePath:    "",
			setup:       nil,
			expectError: true,
		},
		{
			name: "invalid directory permissions",
			setup: func() (string, func()) {
				tempDir := t.TempDir()
				readOnlyDir := filepath.Join(tempDir, "readonly")
				os.Mkdir(readOnlyDir, 0444)
				return filepath.Join(readOnlyDir, "sessions.jsonl"), func() {
					os.Chmod(readOnlyDir, 0755)
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := storage.NewFileStorage(tt.filePath)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, fs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fs)
				assert.FileExists(t, tt.filePath)
			}
		})
	}
}

func TestFileStorage_Save_ErrorCases(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test.jsonl")
	fs, err := storage.NewFileStorage(tempFile)
	require.NoError(t, err)

	os.Chmod(tempFile, 0444)

	err = fs.Save(&models.Session{Task: "test"})
	assert.Error(t, err)

	os.Chmod(tempFile, 0644)

	err = fs.Save(nil)
	assert.Error(t, err)
}

func TestFileStorage_GetAll_EmptyFile(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	sessions, err := fs.GetAll()
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestFileStorage_GetAll_InvalidJSON(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	err := os.WriteFile(filePath, []byte("{invalid json}\n"), 0644)
	require.NoError(t, err)

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	sessions, err := fs.GetAll()
	require.NoError(t, err)
	assert.Empty(t, sessions, "Should skip invalid JSON lines")
}

func TestFileStorage_GetAll_ReadError(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sessions.jsonl")

	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	err = os.Chmod(filePath, 0222)
	require.NoError(t, err)

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	_, err = fs.GetAll()
	require.Error(t, err, "Should return error when file cannot be read")
}

func TestFileStorage_SaveAndGetAll(t *testing.T) {
	tests := []struct {
		name     string
		sessions []*models.Session
		expected int
	}{
		{
			name: "single session",
			sessions: []*models.Session{
				{Task: "task 1", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
			},
			expected: 1,
		},
		{
			name: "multiple sessions",
			sessions: []*models.Session{
				{Task: "task 1", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
				{Task: "task 2", StartTime: time.Now().Add(2 * time.Hour), EndTime: time.Now().Add(3 * time.Hour)},
				{Task: "task 3", StartTime: time.Now().Add(4 * time.Hour), EndTime: time.Now().Add(5 * time.Hour)},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, cleanup := setupTestFile(t)
			defer cleanup()

			fs, err := storage.NewFileStorage(filePath)
			require.NoError(t, err)

			for _, s := range tt.sessions {
				err := fs.Save(s)
				require.NoError(t, err, "Failed to save session")
			}

			allSessions, err := fs.GetAll()
			require.NoError(t, err)
			require.Len(t, allSessions, tt.expected)

			for i, expected := range tt.sessions {
				assert.Equal(t, expected.Task, allSessions[i].Task)
				assert.WithinDuration(t, expected.StartTime, allSessions[i].StartTime, time.Second)
				assert.WithinDuration(t, expected.EndTime, allSessions[i].EndTime, time.Second)
			}
		})
	}
}

func TestFileStorage_GetLast_Empty(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	last, err := fs.GetLast()
	assert.ErrorIs(t, err, models.ErrNoSessions)
	assert.Nil(t, last)
}

func TestFileStorage_GetLast(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	now := time.Now()
	session1 := &models.Session{
		Task:      "test task 1",
		StartTime: now,
		EndTime:   now.Add(time.Hour),
	}

	session2 := &models.Session{
		Task:      "test task 2",
		StartTime: now.Add(2 * time.Hour),
		EndTime:   now.Add(3 * time.Hour),
	}

	err = fs.Save(session1)
	require.NoError(t, err)
	err = fs.Save(session2)
	require.NoError(t, err)

	last, err := fs.GetLast()
	require.NoError(t, err)
	require.NotNil(t, last)

	assert.Equal(t, session2.Task, last.Task)
	assert.WithinDuration(t, session2.StartTime, last.StartTime, time.Second)
	assert.WithinDuration(t, session2.EndTime, last.EndTime, time.Second)
}

func TestFileStorage_GetByDateRange_Invalid(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	sessions, err := fs.GetByDateRange(time.Now().Add(24*time.Hour), time.Now())
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestFileStorage_GetByDateRange(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	emptySessions, err := fs.GetByDateRange(time.Now(), time.Now())
	require.NoError(t, err)
	assert.Empty(t, emptySessions)

	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	testSessions := []models.Session{
		{Task: "task 1", StartTime: now, EndTime: now.Add(time.Hour)},
		{Task: "task 2", StartTime: now.Add(24 * time.Hour), EndTime: now.Add(25 * time.Hour)},
		{Task: "task 3", StartTime: now.Add(48 * time.Hour), EndTime: now.Add(49 * time.Hour)},
	}

	for i := range testSessions {
		err = fs.Save(&testSessions[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		start         time.Time
		end           time.Time
		expectedCount int
		expectedTasks []string
		expectedError bool
	}{
		{
			name:          "single day match",
			start:         now.Add(23 * time.Hour),
			end:           now.Add(25 * time.Hour),
			expectedCount: 1,
			expectedTasks: []string{"task 2"},
		},
		{
			name:          "multiple days match",
			start:         now,
			end:           now.Add(25 * time.Hour),
			expectedCount: 2,
			expectedTasks: []string{"task 1", "task 2"},
		},
		{
			name:          "no matches",
			start:         now.Add(72 * time.Hour),
			end:           now.Add(96 * time.Hour),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fs.GetByDateRange(tt.start, tt.end)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Len(t, result, tt.expectedCount)

				for _, s := range result {
					assert.Contains(t, tt.expectedTasks, s.Task)
				}
			}
		})
	}
}

func TestFileStorage_GetByTask_Empty(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	emptySessions, err := fs.GetByTask("test")
	require.NoError(t, err)
	assert.Empty(t, emptySessions)
}

func TestFileStorage_GetByTask(t *testing.T) {
	filePath, cleanup := setupTestFile(t)
	defer cleanup()

	fs, err := storage.NewFileStorage(filePath)
	require.NoError(t, err)

	now := time.Now()
	testSessions := []models.Session{
		{Task: "test task 1", StartTime: now, EndTime: now.Add(time.Hour)},
		{Task: "test task 2", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)},
		{Task: "test task 1", StartTime: now.Add(4 * time.Hour), EndTime: now.Add(5 * time.Hour)},
	}

	for i := range testSessions {
		err = fs.Save(&testSessions[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		task          string
		expectedCount int
	}{
		{
			name:          "task with multiple sessions",
			task:          "test task 1",
			expectedCount: 2,
		},
		{
			name:          "task with single session",
			task:          "test task 2",
			expectedCount: 1,
		},
		{
			name:          "non-existent task",
			task:          "non-existent",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions, err := fs.GetByTask(tt.task)
			require.NoError(t, err)
			assert.Len(t, sessions, tt.expectedCount)

			for _, s := range sessions {
				assert.Equal(t, tt.task, s.Task)
			}
		})
	}
}
