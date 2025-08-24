package analytics

import (
	"sort"
	"time"

	"github.com/AndriyBarskyi/gotrack/internal/models"
)

const (
	hoursInDay = 24
	
	maxHoursForPerfectScore = 100.0
	maxDaysForPerfectConsistency = 30.0
	maxStreakForPerfectScore = 100.0
	
	hoursWeight = 0.4
	consistencyWeight = 0.4
	streakWeight = 0.2
	
	maxProductivityScore = 100.0
)

// CalculateTotalDuration returns the total duration of all sessions.
func CalculateTotalDuration(ssns []models.Session, task string) time.Duration {
	var totalDuration time.Duration
	for _, ssn := range ssns {
		if task == "" || ssn.Task == task {
			totalDuration += ssn.EndTime.Sub(ssn.StartTime)
		}
	}
	return totalDuration
}

// CalculateTodayDuration returns the total duration of all sessions that started today.
func CalculateTodayDuration(ssns []models.Session, task string) time.Duration {
	var todayDuration time.Duration
	today := time.Now().Format("2006-01-02")
	for _, ssn := range ssns {
		if ssn.StartTime.Format("2006-01-02") == today && (task == "" || ssn.Task == task) {
			todayDuration += ssn.EndTime.Sub(ssn.StartTime)
		}
	}
	return todayDuration
}

// CalculateConsecutiveDays returns the number of recent consecutive days of tracking.
func CalculateConsecutiveDays(ssns []models.Session) int {
	if len(ssns) == 0 {
		return 0
	}

	sorted := make([]models.Session, len(ssns))
	copy(sorted, ssns)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime.After(sorted[j].StartTime)
	})

	consecutiveDays := 1
	currentDate := sorted[0].StartTime.Truncate(hoursInDay * time.Hour)

	for i := 1; i < len(sorted); i++ {
		sessionDate := sorted[i].StartTime.Truncate(hoursInDay * time.Hour)
		if currentDate.Sub(sessionDate) == hoursInDay*time.Hour {
			consecutiveDays++
			currentDate = sessionDate
		} else if currentDate.Sub(sessionDate) > hoursInDay*time.Hour {
			break
		}
	}

	return consecutiveDays
}

// CalculateWeeklyDuration returns the total duration for the current week
func CalculateWeeklyDuration(ssns []models.Session, task string) time.Duration {
	var weeklyDuration time.Duration
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	
	for _, ssn := range ssns {
		if ssn.StartTime.After(startOfWeek) && (task == "" || ssn.Task == task) {
			weeklyDuration += ssn.EndTime.Sub(ssn.StartTime)
		}
	}
	return weeklyDuration
}

// CalculateMonthlyDuration returns the total duration for the current month
func CalculateMonthlyDuration(ssns []models.Session, task string) time.Duration {
	var monthlyDuration time.Duration
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	for _, ssn := range ssns {
		if ssn.StartTime.After(startOfMonth) && (task == "" || ssn.Task == task) {
			monthlyDuration += ssn.EndTime.Sub(ssn.StartTime)
		}
	}
	return monthlyDuration
}

// CalculateYearlyDuration returns the total duration for the current year
func CalculateYearlyDuration(ssns []models.Session, task string) time.Duration {
	var yearlyDuration time.Duration
	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	
	for _, ssn := range ssns {
		if ssn.StartTime.After(startOfYear) && (task == "" || ssn.Task == task) {
			yearlyDuration += ssn.EndTime.Sub(ssn.StartTime)
		}
	}
	return yearlyDuration
}

// GetTopTasks returns the most worked on tasks with their durations
func GetTopTasks(ssns []models.Session, limit int) []TaskStats {
	taskDurations := make(map[string]time.Duration)
	
	for _, ssn := range ssns {
		if !ssn.EndTime.IsZero() {
			taskDurations[ssn.Task] += ssn.EndTime.Sub(ssn.StartTime)
		}
	}
	
	var stats []TaskStats
	for task, duration := range taskDurations {
		stats = append(stats, TaskStats{
			Task:     task,
			Duration: duration,
		})
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Duration > stats[j].Duration
	})
	
	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}
	
	return stats
}

// TaskStats represents statistics for a specific task
type TaskStats struct {
	Task     string
	Duration time.Duration
}

// CalculateLongestStreak returns the longest consecutive days streak in history
func CalculateLongestStreak(ssns []models.Session) int {
	if len(ssns) == 0 {
		return 0
	}
	
	daySet := make(map[string]bool)
	for _, ssn := range ssns {
		day := ssn.StartTime.Format("2006-01-02")
		daySet[day] = true
	}
	
	var days []time.Time
	for dayStr := range daySet {
		day, _ := time.Parse("2006-01-02", dayStr)
		days = append(days, day)
	}
	
	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})
	
	if len(days) == 0 {
		return 0
	}
	
	maxStreak := 1
	currentStreak := 1
	
	for i := 1; i < len(days); i++ {
		if days[i].Sub(days[i-1]) == hoursInDay*time.Hour {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}
	
	return maxStreak
}

// GetProductivityScore calculates a productivity score based on consistency and volume
func GetProductivityScore(ssns []models.Session) float64 {
	if len(ssns) == 0 {
		return 0.0
	}
	
	totalDuration := CalculateTotalDuration(ssns, "")
	consecutiveDays := CalculateConsecutiveDays(ssns)
	longestStreak := CalculateLongestStreak(ssns)
	
	hoursScore := float64(totalDuration.Hours()) / maxHoursForPerfectScore
	consistencyScore := float64(consecutiveDays) / maxDaysForPerfectConsistency
	streakScore := float64(longestStreak) / maxStreakForPerfectScore
	
	if hoursScore > 1.0 {
		hoursScore = 1.0
	}
	if consistencyScore > 1.0 {
		consistencyScore = 1.0
	}
	if streakScore > 1.0 {
		streakScore = 1.0
	}
	
	return (hoursScore*hoursWeight + consistencyScore*consistencyWeight + streakScore*streakWeight) * maxProductivityScore
}
