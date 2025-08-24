package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/AndriyBarskyi/gotrack/internal/models"
	"github.com/AndriyBarskyi/gotrack/internal/tracker"
	"github.com/AndriyBarskyi/gotrack/internal/tracker/analytics"
)

const defaultAmount = 7

type showCmd struct {
	sessionManager *tracker.SessionManager
	amount         int
	today          bool
	task           string
	weekly         bool
	monthly        bool
	yearly         bool
	all            bool
	top            bool
}

// NewShowCmd creates a new show command
func NewShowCmd(sm *tracker.SessionManager) *cobra.Command {
	c := &showCmd{
		sessionManager: sm,
	}

	cmd := &cobra.Command{
		Use:   "show [amount]",
		Short: "Shows all sessions or a specified amount of them",
		Long: `A command-line tool that can be used to show all sessions or a specified amount of them,
	sessions for a specific task, or today's sessions.`,
		Example: `
  gotrack show
  gotrack show 5
  gotrack show --today
  gotrack show --task <task name>
  gotrack show --weekly
  gotrack show --monthly
  gotrack show --all
  gotrack show --top
`,
		Args: cobra.MaximumNArgs(1),
		RunE: c.run,
	}

	cmd.Flags().BoolVarP(&c.today, "today", "t", false, "Show today's sessions")
	cmd.Flags().StringVar(&c.task, "task", "", "Show sessions for a specific task")
	cmd.Flags().BoolVarP(&c.weekly, "weekly", "w", false, "Show weekly statistics")
	cmd.Flags().BoolVarP(&c.monthly, "monthly", "m", false, "Show monthly statistics")
	cmd.Flags().BoolVarP(&c.yearly, "yearly", "y", false, "Show yearly statistics")
	cmd.Flags().BoolVar(&c.all, "all", false, "Show comprehensive statistics")
	cmd.Flags().BoolVar(&c.top, "top", false, "Show top tasks by time spent")

	return cmd
}

func (c *showCmd) run(cmd *cobra.Command, args []string) error {
	sm := c.sessionManager
	if sm == nil {
		sm = GetSessionManager()
		if sm == nil {
			fmt.Println("No session manager available. Please ensure GoTrack is properly initialized.")
			return fmt.Errorf("session manager not initialized")
		}
	}

	var ssns []models.Session
	var err error

	if c.today {
		ssns, err = sm.GetTodaySessions()
	} else if c.task != "" {
		ssns, err = sm.GetSessionsForTask(c.task)
	} else {
		ssns, err = sm.GetAllSessions()
	}

	if err != nil {
		return fmt.Errorf("failed to get sessions: %v", err)
	}

	if len(args) > 0 {
		_, err := fmt.Sscanf(args[0], "%d", &c.amount)
		if err != nil || c.amount <= 0 {
			c.amount = defaultAmount
		}
	}

	if c.amount > len(ssns) {
		c.amount = len(ssns)
	}

	for i := len(ssns) - 1; i >= len(ssns)-c.amount && i >= 0; i-- {
		fmt.Println(sm.FormatSession(ssns[i], i, ssns))
	}

	if len(ssns) > 0 {
		todayDuration := analytics.CalculateTodayDuration(ssns, c.task)
		totalDuration := analytics.CalculateTotalDuration(ssns, c.task)

		if c.task != "" {
			fmt.Printf("Today duration for task %s: %s\n",
				color.CyanString(c.task),
				formatDuration(todayDuration))
			fmt.Printf("Total duration for task %s: %s\n",
				color.CyanString(c.task),
				formatDuration(totalDuration))
		} else {
			fmt.Printf("Today duration: %s\n", formatDuration(todayDuration))
			fmt.Printf("Total duration: %s\n", formatDuration(totalDuration))

			if c.weekly || c.all {
				weeklyDuration := analytics.CalculateWeeklyDuration(ssns, c.task)
				fmt.Printf("Weekly duration: %s\n", formatDuration(weeklyDuration))
			}

			if c.monthly || c.all {
				monthlyDuration := analytics.CalculateMonthlyDuration(ssns, c.task)
				fmt.Printf("Monthly duration: %s\n", formatDuration(monthlyDuration))
			}

			if c.yearly || c.all {
				yearlyDuration := analytics.CalculateYearlyDuration(ssns, c.task)
				fmt.Printf("Yearly duration: %s\n", formatDuration(yearlyDuration))
			}
		}

		fmt.Printf("Consecutive days: %d\n", analytics.CalculateConsecutiveDays(ssns))

		if c.all {
			longestStreak := analytics.CalculateLongestStreak(ssns)
			productivityScore := analytics.GetProductivityScore(ssns)
			fmt.Printf("Longest streak: %d days\n", longestStreak)
			fmt.Printf("Productivity score: %.1f/100\n", productivityScore)
		}

		if c.top || c.all {
			fmt.Println("\nTop Tasks:")
			topTasks := analytics.GetTopTasks(ssns, 5)
			for i, task := range topTasks {
				fmt.Printf("%d. %s: %s\n", i+1,
					color.CyanString(task.Task),
					formatDuration(task.Duration))
			}
		}
	} else {
		fmt.Println("No sessions found")
	}

	return nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
