package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AndriyBarskyi/gotrack/internal/config"
	"github.com/AndriyBarskyi/gotrack/internal/storage"
	"github.com/AndriyBarskyi/gotrack/internal/tracker"
)

var (
	appConfig      *config.Config
	sessionManager *tracker.SessionManager
	sessionStorage storage.Storage
)

var rootCmd = &cobra.Command{
	Use:   "gotrack",
	Short: "A time tracking CLI tool",
	Long: `A command-line tool for tracking time spent on tasks.

Track your time with ease using simple commands. Get started by creating a new
session with 'gotrack start' and stop it with 'gotrack stop'.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(NewStartCmd(nil))
	rootCmd.AddCommand(NewStopCmd(nil))
	rootCmd.AddCommand(NewShowCmd(nil))
	rootCmd.AddCommand(NewCurrentCmd(nil))
	rootCmd.AddCommand(NewPomoCmd(nil))
	rootCmd.AddCommand(NewStatusCmd(nil))
}

// GetSessionManager returns the initialized session manager
func GetSessionManager() *tracker.SessionManager {
	return sessionManager
}

// initConfig loads the application configuration
func initConfig() {
	var err error
	appConfig, err = config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	gotrackDir := filepath.Join(homeDir, ".gotrack")
	if err := os.MkdirAll(gotrackDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .gotrack directory: %v\n", err)
		os.Exit(1)
	}

	sessionStorage, err = storage.NewFileStorage(filepath.Join(gotrackDir, "sessions.jsonl"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	sessionManager = tracker.NewSessionManager(sessionStorage)
}
