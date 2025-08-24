package pomodoro

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/config"
)

// Callback functions type
type (
	StateChangeFunc func(State)
	TickFunc        func(remaining time.Duration)
)

// Pomodoro represents a Pomodoro timer instance
type Pomodoro struct {
	config       *config.PomodoroConfig
	state        State
	remaining    time.Duration
	cycles       int
	workSessions int

	ticker   *time.Ticker
	tickerQuit chan struct{}
	lastTick time.Time
	mu       sync.Mutex

	onStateChange StateChangeFunc
	onTick        TickFunc
}

// Config returns the Pomodoro configuration
func (p *Pomodoro) Config() *config.PomodoroConfig {
	return p.config
}

// ErrAlreadyRunning is returned when trying to start an already running Pomodoro
var ErrAlreadyRunning = errors.New("pomodoro is already running")

// New creates a new Pomodoro timer with the given configuration
func New(cfg *config.PomodoroConfig) *Pomodoro {
	return &Pomodoro{
		config:        cfg,
		state:         StateIdle,
		remaining:     cfg.WorkDuration,
		onStateChange: func(State) {},
		onTick:        func(time.Duration) {},
	}
}

// OnStateChange sets the callback for state changes
func (p *Pomodoro) OnStateChange(fn StateChangeFunc) {
	p.onStateChange = fn
}

// OnTick sets the callback for timer ticks
func (p *Pomodoro) OnTick(fn TickFunc) {
	p.onTick = fn
}

// Start starts the Pomodoro timer
func (p *Pomodoro) Start() error {
	p.mu.Lock()

	if p.state == StateWorking || p.state == StateShortBreak || p.state == StateLongBreak {
		p.mu.Unlock()
		return fmt.Errorf("cannot start: timer is already running")
	}

	if p.state == StatePaused {
		p.state = StateWorking
	} else {
		p.remaining = p.config.WorkDuration
		p.state = StateWorking
	}
	newState := p.state
	p.mu.Unlock()
	if p.onStateChange != nil {
		p.onStateChange(newState)
	}
	p.startTicker()

	return nil
}

// Pause pauses the Pomodoro timer
func (p *Pomodoro) Pause() {
	p.mu.Lock()
	if p.state != StateWorking && p.state != StateShortBreak && p.state != StateLongBreak {
		p.mu.Unlock()
		return
	}

	if p.ticker != nil {
		p.ticker.Stop()
	}

	p.state = StatePaused
	newState := p.state
	p.mu.Unlock()
	if p.onStateChange != nil {
		p.onStateChange(newState)
	}
}

// Stop stops the Pomodoro timer
func (p *Pomodoro) Stop() {
	p.mu.Lock()
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}
	if p.tickerQuit != nil {
		close(p.tickerQuit)
		p.tickerQuit = nil
	}
	p.state = StateIdle
	p.cycles = 0
	p.workSessions = 0
	p.lastTick = time.Time{}
	newState := p.state
	p.mu.Unlock()
	if p.onStateChange != nil {
		p.onStateChange(newState)
	}
}

// State returns the current state of the Pomodoro timer
func (p *Pomodoro) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Cycles returns the number of completed work sessions
func (p *Pomodoro) Cycles() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cycles
}

// Remaining returns the remaining time in the current session
func (p *Pomodoro) Remaining() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.remaining
}

func (p *Pomodoro) startTicker() {
	p.mu.Lock()
	if p.ticker != nil {
		p.ticker.Stop()
	}
	if p.tickerQuit != nil {
		close(p.tickerQuit)
		p.tickerQuit = nil
	}

	p.ticker = time.NewTicker(100 * time.Millisecond)
	p.tickerQuit = make(chan struct{})
	p.lastTick = time.Now()
	p.mu.Unlock()

	go func(localTicker *time.Ticker, quit <-chan struct{}) {
		for {
			select {
			case <-localTicker.C:
				p.mu.Lock()
				if p.state != StateWorking && p.state != StateShortBreak && p.state != StateLongBreak {
					p.mu.Unlock()
					continue
				}
				p.remaining -= 100 * time.Millisecond
				remaining := p.remaining
				shouldContinue := false
				if remaining <= 0 {
					shouldContinue = p.state == StateWorking ||
						(p.state == StateShortBreak && p.config.AutoStartBreak) ||
						(p.state == StateLongBreak && p.config.AutoStartBreak)
				}
				p.mu.Unlock()

				if p.onTick != nil {
					p.onTick(remaining)
				}

				if remaining <= 0 {
					p.completeSession()
					if !shouldContinue {
						return
					}
				}
			case <-quit:
				return
			}
		}
	}(p.ticker, p.tickerQuit)
}

func (p *Pomodoro) completeSession() {

	p.mu.Lock()
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}
	if p.tickerQuit != nil {
		close(p.tickerQuit)
		p.tickerQuit = nil
	}

	switch p.state {
	case StateWorking:
		p.workSessions++
		p.cycles++

		if p.workSessions > 0 && p.workSessions%p.config.LongBreakInterval == 0 {
			p.remaining = p.config.LongBreak
			p.state = StateLongBreak
		} else {
			p.remaining = p.config.BreakDuration
			p.state = StateShortBreak
		}

	case StateShortBreak, StateLongBreak:
		p.remaining = p.config.WorkDuration
		p.state = StateWorking
	}

	newState := p.state
	autoStart := p.config.AutoStartBreak || p.state == StateWorking
	p.mu.Unlock()

	if p.onStateChange != nil {
		p.onStateChange(newState)
	}

	if autoStart {
		p.startTicker()
	}
}
