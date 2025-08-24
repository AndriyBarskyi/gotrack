package pomodoro

// State represents the current state of the Pomodoro timer
type State int

const (
	// StateIdle means the timer is not running
	StateIdle State = iota
	// StateWorking means a work session is active
	StateWorking
	// StateShortBreak means a short break is active
	StateShortBreak
	// StateLongBreak means a long break is active
	StateLongBreak
	// StatePaused means the timer is paused
	StatePaused
)

// String returns a human-readable representation of the state
func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateWorking:
		return "working"
	case StateShortBreak:
		return "short break"
	case StateLongBreak:
		return "long break"
	case StatePaused:
		return "paused"
	default:
		return "unknown"
	}
}
