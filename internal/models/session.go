package models

import (
	"errors"
	"time"
)

// ErrNoSessions is returned when no sessions are found
var ErrNoSessions = errors.New("no sessions found")

// Session represents a work session
type Session struct {
	Task      string    `json:"task"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// IsActive returns true if the session is currently active (started but not finished)
func (s *Session) IsActive() bool {
	return !s.StartTime.IsZero() && s.EndTime.IsZero()
}

// Duration returns the duration of the session.
// If the session hasn't started (StartTime is zero), it returns 0.
// If the session is in progress (EndTime is zero), it returns the duration from StartTime to now.
// If the session is completed, it returns the duration between StartTime and EndTime.
func (s *Session) Duration() time.Duration {
	if s.StartTime.IsZero() {
		return 0
	}
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}
