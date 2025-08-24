package models_test

import (
	"testing"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSession_IsActive(t *testing.T) {
	tests := []struct {
		name    string
		session models.Session
		want    bool
	}{
		{
			name:    "not started",
			session: models.Session{},
			want:    false,
		},
		{
			name: "active session",
			session: models.Session{
				Task:      "test",
				StartTime: time.Now(),
			},
			want: true,
		},
		{
			name: "completed session",
			session: models.Session{
				Task:      "test",
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   time.Now(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.IsActive()
			assert.Equal(t, tt.want, got, "IsActive() = %v, want %v", got, tt.want)
		})
	}
}

func TestSession_Duration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		session models.Session
		want    time.Duration
	}{
		{
			name:    "not started",
			session: models.Session{},
			want:    0,
		},
		{
			name: "active session",
			session: models.Session{
				StartTime: now.Add(-30 * time.Minute),
			},
			want: 30 * time.Minute,
		},
		{
			name: "completed session",
			session: models.Session{
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(-time.Hour),
			},
			want: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tt.session.Duration()

			switch tt.name {
			case "active session":
				assert.GreaterOrEqual(t, duration, tt.want, "Duration() should be at least %v", tt.want)
			case "not started":
				assert.Zero(t, duration, "Duration() should be 0 for not started sessions")
			default:
				assert.Equal(t, tt.want, duration, "Duration() = %v, want %v", duration, tt.want)
			}
		})
	}
}
