package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndriyBarskyi/gotrack/internal/tracker"
)

type currentCmd struct {
	sessionManager *tracker.SessionManager
}

// NewCurrentCmd creates a new current command
func NewCurrentCmd(sm *tracker.SessionManager) *cobra.Command {
	c := &currentCmd{
		sessionManager: sm,
	}
	return &cobra.Command{
		Use:   "current",
		Short: "Show the current session time dynamically",
		Long: `Show a live-updating timer for the current active session.
The display updates every second until interrupted with Ctrl+C.`,
		Args: cobra.NoArgs,
		RunE: c.run,
	}
}

func (c *currentCmd) run(cmd *cobra.Command, args []string) error {
	sm := c.sessionManager
	if sm == nil {
		sm = GetSessionManager()
		if sm == nil {
			fmt.Println("No session manager available. Please ensure GoTrack is properly initialized.")
			fmt.Println("This might happen if the configuration failed to load.")
			return fmt.Errorf("session manager not initialized")
		}
	}

	session, err := sm.GetLast()
	if err != nil {
		fmt.Println("No sessions found. Start a session with 'gotrack start <task>'.")
		return nil
	}
	
	if session == nil {
		fmt.Println("No sessions found. Start a session with 'gotrack start <task>'.")
		return nil
	}
	
	if !session.EndTime.IsZero() {
		fmt.Println("No active session found. The last session has already ended.")
		fmt.Printf("Last session: %s (ended at %s)\n", 
			session.Task, 
			session.EndTime.Format("15:04:05"))
		return nil
	}

	fmt.Printf("Tracking current session: %s\n", session.Task)
	fmt.Println("Press Ctrl+C to stop monitoring...")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			fmt.Println("\nStopped tracking current session.")
			return nil
		case <-ticker.C:
			currentSession, err := sm.GetLast()
			if err != nil || currentSession == nil {
				fmt.Println("\nSession ended.")
				return nil
			}
			
			if !currentSession.EndTime.IsZero() {
				fmt.Println("\nSession ended.")
				return nil
			}

			duration := time.Since(currentSession.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			seconds := int(duration.Seconds()) % 60

			fmt.Printf("\rFocusing on: %s | %02d:%02d:%02d",
				currentSession.Task, hours, minutes, seconds)
		}
	}
}
