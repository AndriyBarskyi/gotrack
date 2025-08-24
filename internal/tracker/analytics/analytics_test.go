package analytics_test

import (
	"testing"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/models"
	"github.com/AndriyBarskyi/gotrack/internal/tracker/analytics"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalDuration(t *testing.T) {
	tests := []struct {
		name     string
		sessions []models.Session
		task     string
		expected time.Duration
	}{
		{
			name:     "empty sessions",
			sessions: []models.Session{},
			task:     "",
			expected: 0,
		},
		{
			name: "single session",
			sessions: []models.Session{
				{
					Task:      "test",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
				},
			},
			task:     "",
			expected: time.Hour,
		},
		{
			name: "filter by task",
			sessions: []models.Session{
				{
					Task:      "test1",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
				},
				{
					Task:      "test2",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC),
				},
			},
			task:     "test1",
			expected: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analytics.CalculateTotalDuration(tt.sessions, tt.task)
			if tt.expected == 0 {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.InEpsilon(t, tt.expected.Seconds(), result.Seconds(), 1e-9)
			}
		})
	}
}

func TestCalculateTodayDuration(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	tests := []struct {
		name     string
		sessions []models.Session
		task     string
		expected time.Duration
	}{
		{
			name: "session from today",
			sessions: []models.Session{
				{
					Task:      "test",
					StartTime: today,
					EndTime:   today.Add(time.Hour),
				},
			},
			task:     "",
			expected: time.Hour,
		},
		{
			name: "session from yesterday",
			sessions: []models.Session{
				{
					Task:      "test",
					StartTime: yesterday,
					EndTime:   yesterday.Add(time.Hour),
				},
			},
			task:     "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analytics.CalculateTodayDuration(tt.sessions, tt.task)
			if tt.expected == 0 {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.InEpsilon(t, tt.expected.Seconds(), result.Seconds(), 1e-9)
			}
		})
	}
}

func TestCalculateConsecutiveDays(t *testing.T) {
	tests := []struct {
		name     string
		sessions []models.Session
		expected int
	}{
		{
			name:     "no sessions",
			sessions: []models.Session{},
			expected: 0,
		},
		{
			name: "one day",
			sessions: []models.Session{
				{
					Task:      "test",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
				},
			},
			expected: 1,
		},
		{
			name: "two consecutive days",
			sessions: []models.Session{
				{
					Task:      "test",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
				},
				{
					Task:      "test",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, -1),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC).AddDate(0, 0, -1),
				},
			},
			expected: 2,
		},
		{
			name: "non-consecutive days",
			sessions: []models.Session{
				{
					Task:      "test",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
				},
				{
					Task:      "test",
					StartTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, -2),
					EndTime:   time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC).AddDate(0, 0, -2),
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analytics.CalculateConsecutiveDays(tt.sessions)
			assert.Equal(t, tt.expected, result)
		})
	}
}
