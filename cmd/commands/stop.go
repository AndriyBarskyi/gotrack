package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/AndriyBarskyi/gotrack/internal/tracker"
)

type stopCmd struct {
	sessionManager *tracker.SessionManager
}

// NewStopCmd creates a new stop command
func NewStopCmd(sm *tracker.SessionManager) *cobra.Command {
	c := &stopCmd{
		sessionManager: sm,
	}
	return &cobra.Command{
		Use:     "stop",
		Short:   "Stop tracking the current task",
		Long:    `Stop tracking the currently running task and record the end time.`,
		Example: `  gotrack stop`,
		Args:    cobra.NoArgs,
		RunE:    c.run,
	}
}

func (c *stopCmd) run(cmd *cobra.Command, args []string) error {
	sm := c.sessionManager
	if sm == nil {
		sm = GetSessionManager()
		if sm == nil {
			fmt.Println("No session manager available. Please ensure GoTrack is properly initialized.")
			return fmt.Errorf("session manager not initialized")
		}
	}

	lastSession, err := sm.GetLast()
	if err != nil {
		return fmt.Errorf("no active session found")
	}

	if lastSession != nil && !lastSession.EndTime.IsZero() {
		return fmt.Errorf("no active session to stop")
	}

	session, err := sm.Finish()
	if err != nil {
		return fmt.Errorf("failed to stop session: %v", err)
	}

	duration := session.EndTime.Sub(session.StartTime).Round(time.Second)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	fmt.Printf("Stopped tracking %s after %02d:%02d:%02d\n",
		color.CyanString(session.Task),
		hours, minutes, seconds,
	)

	return nil
}
