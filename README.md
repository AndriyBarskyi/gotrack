# GoTrack

A command-line time tracking and productivity tool with Pomodoro timer support.

## Features

- **Time Tracking**: Start and stop tracking tasks with simple commands
- **Pomodoro Timer**: Built-in Pomodoro technique support with customizable work/break intervals
- **Analytics**: View daily, weekly, and monthly statistics
- **Session Management**: Automatic validation to prevent overlapping sessions
- **Streak Tracking**: Monitor consecutive working days
- **Task Statistics**: Detailed breakdown of time spent per task

## Installation

```bash
# Build from source
go build -o gotrack ./cmd/gotrack
```

## Quick Start

```bash
# Start tracking a task
gotrack start "Working on feature X"

# Check current status
gotrack current

# Stop tracking
gotrack stop

# View today's summary
gotrack show

# Start a Pomodoro session
gotrack pomo start "Deep work session"
```

## Commands

### Basic Time Tracking

- `gotrack start <task>` - Start tracking a new task
- `gotrack stop` - Stop the current tracking session
- `gotrack current` - Show currently active session with live timer
- `gotrack status` - Quick status check

### Analytics & Reports

- `gotrack show` - Show today's sessions and statistics
- `gotrack show --task <name>` - Show statistics for a specific task
- `gotrack show --all` - Show all-time statistics

### Pomodoro Timer

- `gotrack pomo start <task>` - Start a Pomodoro session
- `gotrack pomo stop` - Stop the current Pomodoro session
- `gotrack pomo status` - Check Pomodoro timer status

## Configuration

GoTrack stores sessions in `~/.gotrack/sessions.jsonl` by default.

### Pomodoro Settings

Default Pomodoro configuration:
- Work duration: 25 minutes
- Short break: 5 minutes
- Long break: 15 minutes
- Long break interval: Every 4 work sessions
- Auto-start breaks: Enabled
- Notifications: Enabled

## Examples

```bash
# Track a meeting
gotrack start "Team standup meeting"
# ... meeting happens ...
gotrack stop

# Work with Pomodoro technique
gotrack pomo start "Code review"
# Timer runs automatically with breaks

# Check your productivity
gotrack show
# Output:
# Today's Sessions:
# Team standup meeting: 00:30:15
# Code review: 01:45:30
# 
# Today duration: 02:15:45
# Consecutive days: 5
```

## Data Storage

Sessions are stored in JSONL format at `~/.gotrack/sessions.jsonl`. Each session contains:
- Task name
- Start time
- End time (when completed)
- Duration calculations

## Contributing

This is a personal productivity tool built with Go. Feel free to fork and customize for your needs.

## License

MIT License