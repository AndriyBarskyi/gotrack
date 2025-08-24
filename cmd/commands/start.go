package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/AndriyBarskyi/gotrack/internal/storage"
	"github.com/AndriyBarskyi/gotrack/internal/tracker"
)

type startCmd struct {
	sessionManager *tracker.SessionManager
}

// NewStartCmd creates a new start command
func NewStartCmd(sm *tracker.SessionManager) *cobra.Command {
	c := &startCmd{
		sessionManager: sm,
	}
	return &cobra.Command{
		Use:   "start <task name>",
		Short: "Start tracking a task",
		Long:  `Start tracking time for a specific task. This will create a new session.`,
		Example: `  gotrack start "Working on feature X"
  gotrack start "Meeting with team"`,
		Args: cobra.ExactArgs(1),
		RunE: c.run,
	}
}

func (c *startCmd) run(cmd *cobra.Command, args []string) error {
	sm := c.sessionManager
	if sm == nil {
		sm = GetSessionManager()
		if sm == nil {
			fmt.Println("No session manager available. Please ensure GoTrack is properly initialized.")
			return fmt.Errorf("session manager not initialized")
		}
	}

	session, err := sm.Start(args[0])
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}

	fmt.Printf("Started tracking %s at %s\n",
		color.CyanString(session.Task),
		session.StartTime.Format("15:04:05"),
	)
	return nil
}

func init() {
	storage, err := storage.NewFileStorage("~/.gotrack/sessions.jsonl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}
	sessionManager := tracker.NewSessionManager(storage)

	rootCmd.AddCommand(NewStartCmd(sessionManager))
}
