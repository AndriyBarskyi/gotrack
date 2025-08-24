package pomodoro_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/AndriyBarskyi/gotrack/internal/config"
	"github.com/AndriyBarskyi/gotrack/internal/tracker/pomodoro"
)

// testConfig returns a default test configuration
func testConfig() *config.PomodoroConfig {
	return &config.PomodoroConfig{
		WorkDuration:      25 * time.Minute,
		BreakDuration:     5 * time.Minute,
		LongBreak:         15 * time.Minute,
		LongBreakInterval: 4,
		AutoStartBreak:    true,
	}
}

// newTestPomodoro creates a new Pomodoro instance with test configuration
func newTestPomodoro() *pomodoro.Pomodoro {
	cfg := testConfig()
	return pomodoro.New(cfg)
}

func TestNew(t *testing.T) {
	p := newTestPomodoro()

	assert.Equal(t, 25*time.Minute, p.Remaining(), "Initial remaining time should match work duration")
	assert.Equal(t, pomodoro.StateIdle, p.State(), "Initial state should be idle")
	assert.Equal(t, 0, p.Cycles(), "Initial cycles should be 0")
}

func TestStopWhenNotRunning(t *testing.T) {
	p := newTestPomodoro()

	assert.NotPanics(t, p.Stop, "Stop should not panic when not running")
}

func TestStart(t *testing.T) {
	t.Run("start from idle", func(t *testing.T) {
		p := newTestPomodoro()

		err := p.Start()

		assert.NoError(t, err)
		assert.Equal(t, pomodoro.StateWorking, p.State())

		p.Stop()
	})

	t.Run("start from paused", func(t *testing.T) {
		p := newTestPomodoro()
		p.Start()
		p.Pause()

		err := p.Start()

		assert.NoError(t, err)
		assert.Equal(t, pomodoro.StateWorking, p.State())
	})

	t.Run("error when already running", func(t *testing.T) {
		p := newTestPomodoro()
		p.Start()

		err := p.Start()

		assert.Error(t, err)
		assert.Equal(t, "cannot start: timer is already running", err.Error())
	})
}

func TestStartStop(t *testing.T) {
	t.Run("normal start/stop", func(t *testing.T) {
		p := newTestPomodoro()

		err := p.Start()
		assert.NoError(t, err)
		assert.Equal(t, pomodoro.StateWorking, p.State())

		p.Stop()
		assert.Equal(t, pomodoro.StateIdle, p.State())
	})

	t.Run("double start", func(t *testing.T) {
		p := newTestPomodoro()

		err := p.Start()
		assert.NoError(t, err)

		err = p.Start()
		assert.Error(t, err, "Second Start() should return an error")
		assert.Equal(t, "cannot start: timer is already running", err.Error())

		p.Stop()
	})
}

func TestPauseResume(t *testing.T) {
	p := newTestPomodoro()

	p.Start()
	p.Pause()
	assert.Equal(t, pomodoro.StatePaused, p.State(), "State should be paused after pause")

	remaining := p.Remaining()
	err := p.Start()
	assert.NoError(t, err, "Resume should not return an error")
	assert.Equal(t, pomodoro.StateWorking, p.State(), "State should be working after resume")
	assert.Equal(t, remaining, p.Remaining(), "Remaining time should be preserved after resume")
}

func TestStateTransitions(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		p := newTestPomodoro()
		assert.Equal(t, pomodoro.StateIdle, p.State())
	})

	t.Run("work session completes and transitions to break", func(t *testing.T) {
		p := newTestPomodoro()

		workDuration := 2 * time.Second
		p.Config().WorkDuration = workDuration
		p.Config().BreakDuration = 1 * time.Second
		p.Config().AutoStartBreak = true

		t.Logf("Starting test with work duration: %v, break duration: %v",
			p.Config().WorkDuration, p.Config().BreakDuration)

		stateCh := make(chan pomodoro.State, 10)

		p.OnStateChange(func(s pomodoro.State) {
			t.Logf("State changed to: %s, remaining: %v", s, p.Remaining())
			stateCh <- s
		})

		p.OnTick(func(d time.Duration) {
			t.Logf("Tick - Remaining: %v, State: %s", d, p.State())
		})

		t.Log("Starting Pomodoro...")
		err := p.Start()
		assert.NoError(t, err)

		t.Log("Waiting for working state...")
		select {
		case state := <-stateCh:
			assert.Equal(t, pomodoro.StateWorking, state, "Should transition to working state")
			t.Logf("In working state, remaining: %v", p.Remaining())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for working state")
		}

		t.Log("Waiting for work session to complete and transition to break...")
		startTime := time.Now()

		select {
		case state := <-stateCh:
			elapsed := time.Since(startTime)
			t.Logf("State changed to %s after %v, remaining: %v", state, elapsed, p.Remaining())

			if state != pomodoro.StateShortBreak {
				t.Fatalf("Expected state to change to short break, got: %s", state)
			}

			assert.Equal(t, p.Config().BreakDuration, p.Remaining(), "Remaining time should be break duration")

		case <-time.After(workDuration + 2*time.Second):
			t.Fatalf("Timed out waiting for short break state after %v. Current state: %s, remaining: %v",
				time.Since(startTime), p.State(), p.Remaining())
		}

		t.Log("Test completed, stopping Pomodoro...")
		p.Stop()
		t.Log("Pomodoro stopped")
	})
}

func TestCallbacks(t *testing.T) {
	t.Run("state change callback", func(t *testing.T) {
		p := newTestPomodoro()

		var stateChanges []pomodoro.State
		p.OnStateChange(func(s pomodoro.State) {
			stateChanges = append(stateChanges, s)
		})

		err := p.Start()
		assert.NoError(t, err)
		assert.Contains(t, stateChanges, pomodoro.StateWorking, "Should have transitioned to working state")

		p.Stop()
	})

	t.Run("tick callback", func(t *testing.T) {
		p := newTestPomodoro()

		p.Config().WorkDuration = 2 * time.Second

		var tickCount int
		var lastRemaining time.Duration

		p.OnTick(func(d time.Duration) {
			tickCount++
			lastRemaining = d
		})

		err := p.Start()
		assert.NoError(t, err)

		time.Sleep(1100 * time.Millisecond)

		p.Stop()

		assert.Greater(t, tickCount, 0, "Should have received tick callbacks")
		assert.Less(t, lastRemaining, p.Config().WorkDuration, "Remaining time should have decreased")
	})
}
