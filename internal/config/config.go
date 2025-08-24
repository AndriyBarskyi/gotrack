package config

import "time"

// Config holds the application configuration
type Config struct {
	Pomodoro PomodoroConfig `yaml:"pomodoro"`
}

// PomodoroConfig holds the configuration for the Pomodoro timer
type PomodoroConfig struct {
	// WorkDuration is the duration of a work session
	WorkDuration time.Duration `yaml:"work_duration"`
	// BreakDuration is the duration of a short break
	BreakDuration time.Duration `yaml:"break_duration"`
	// LongBreak is the duration of a long break
	LongBreak time.Duration `yaml:"long_break"`
	// LongBreakInterval is the number of work sessions before a long break
	LongBreakInterval int `yaml:"long_break_interval"`
	// AutoStartBreak whether to auto-start the next break
	AutoStartBreak bool `yaml:"auto_start_break"`
}

// Default returns the default application configuration
func Default() *Config {
	return &Config{
		Pomodoro: PomodoroConfig{
			WorkDuration:     25 * time.Minute,
			BreakDuration:    5 * time.Minute,
			LongBreak:        15 * time.Minute,
			LongBreakInterval: 4,
			AutoStartBreak:   true,
		},
	}
}
