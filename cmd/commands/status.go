package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AndriyBarskyi/gotrack/internal/tracker"
)

type statusCmd struct {
	sessionManager *tracker.SessionManager
}

// NewStatusCmd creates a new status command
func NewStatusCmd(sm *tracker.SessionManager) *cobra.Command {
	c := &statusCmd{
		sessionManager: sm,
	}

	return &cobra.Command{
		Use:   "status",
		Short: "Show the status of the last session",
		Args:  cobra.NoArgs,
		RunE:  c.run,
	}
}

func (c *statusCmd) run(cmd *cobra.Command, args []string) error {
	sm := c.sessionManager
	if sm == nil {
		sm = GetSessionManager()
		if sm == nil {
			fmt.Println("No session manager available. Please ensure GoTrack is properly initialized.")
			return fmt.Errorf("session manager not initialized")
		}
	}

	session, err := sm.GetLast()
	if err != nil {
		return fmt.Errorf("failed to get last session: %v", err)
	}

	if session == nil {
		fmt.Println("No sessions found")
		return nil
	}

	sessions, err := sm.GetAllSessions()
	if err != nil {
		return fmt.Errorf("failed to get sessions: %v", err)
	}

	fmt.Println(sm.FormatSession(*session, 0, sessions))
	return nil
}
