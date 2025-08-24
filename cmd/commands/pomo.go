package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/AndriyBarskyi/gotrack/internal/config"
	"github.com/AndriyBarskyi/gotrack/internal/tracker"
	pkgPomodoro "github.com/AndriyBarskyi/gotrack/internal/tracker/pomodoro"
)

type pomoCmd struct {
	sessionManager *tracker.SessionManager
	workDuration   time.Duration
	breakDuration  time.Duration
	cycles         int
}

// NewPomoCmd creates a new pomodoro command
func NewPomoCmd(sm *tracker.SessionManager) *cobra.Command {
	cmd := &pomoCmd{
		sessionManager: sm,
	}

	cobraCmd := &cobra.Command{
		Use:   "pomo [task name]",
		Short: "Start a Pomodoro timer for a task",
		Long: `Start a Pomodoro timer with work and break intervals.

By default, it runs for 25 minutes of work followed by 5 minutes of break.
You can customize the durations using the flags.`,
		Example: `
  # Start a default Pomodoro (25m work, 5m break)
  gotrack pomo "Coding"

  # Custom work and break durations
  gotrack pomo "Writing" --work 50m --break 10m

  # Run multiple cycles
  gotrack pomo "Studying" --cycles 4
`,
		Args: cobra.ExactArgs(1),
		RunE: cmd.run,
	}

	cobraCmd.Flags().DurationVarP(&cmd.workDuration, "work", "w", cfg.Default().Pomodoro.WorkDuration, "Work duration")
	cobraCmd.Flags().DurationVarP(&cmd.breakDuration, "break", "b", cfg.Default().Pomodoro.BreakDuration, "Break duration")
	cobraCmd.Flags().IntVarP(&cmd.cycles, "cycles", "c", 1, "Number of work/break cycles")

	return cobraCmd
}

func (c *pomoCmd) run(cmd *cobra.Command, args []string) error {
	taskName := args[0]

	sm := c.sessionManager
	if sm == nil {
		sm = GetSessionManager()
		if sm == nil {
			fmt.Println("No session manager available. Please ensure GoTrack is properly initialized.")
			return fmt.Errorf("session manager not initialized")
		}
	}

	pomodoroCfg := appConfig.Pomodoro
	if c.workDuration != cfg.Default().Pomodoro.WorkDuration {
		pomodoroCfg.WorkDuration = c.workDuration
	}
	if c.breakDuration != cfg.Default().Pomodoro.BreakDuration {
		pomodoroCfg.BreakDuration = c.breakDuration
	}

	pomodoro := pkgPomodoro.New(&pomodoroCfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if _, err := sm.Start(taskName); err != nil {
		return fmt.Errorf("failed to start work session: %v", err)
	}

	pomodoro.OnStateChange(func(s pkgPomodoro.State) {
		switch s {
		case pkgPomodoro.StateWorking:
			fmt.Printf("\n\nStarting work session\n")
		case pkgPomodoro.StateShortBreak, pkgPomodoro.StateLongBreak:
			fmt.Printf("\n\nStarting %s\n", s.String())
		}
	})

	fmt.Printf("Press Ctrl+C to stop the session...")

	if err := pomodoro.Start(); err != nil {
		return fmt.Errorf("failed to start Pomodoro: %v", err)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\nStopping Pomodoro session...")
			pomodoro.Stop()
			if _, err := sm.Finish(); err != nil {
				fmt.Printf("Error finishing session: %v\n", err)
			}
			return nil
		case <-ticker.C:
			state := pomodoro.State()
			if state == pkgPomodoro.StateIdle {
				fmt.Println("\nPomodoro session completed!")
				if _, err := sm.Finish(); err != nil {
					fmt.Printf("Error finishing session: %v\n", err)
				}
				return nil
			}

			remaining := max(pomodoro.Remaining(), 0)

			hours := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60
			seconds := int(remaining.Seconds()) % 60

			stateStr := ""
			switch state {
			case pkgPomodoro.StateWorking:
				stateStr = "Work"
			case pkgPomodoro.StateShortBreak:
				stateStr = "Short Break"
			case pkgPomodoro.StateLongBreak:
				stateStr = "Long Break"
			case pkgPomodoro.StatePaused:
				stateStr = "Paused"
			default:
				stateStr = state.String()
			}

			fmt.Printf("\r%s: %s | %02d:%02d:%02d", stateStr, taskName, hours, minutes, seconds)
			os.Stdout.Sync()
		}
	}
}
