package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tasklib "brooce/task"
)

// Example cron entry:
// SET "brooce:cron:jobs:schedule_postprocess" "* * * * * queue:common gorunjob schedule_postprocess"

type CronType struct {
	Name string

	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string

	Queue         string
	Command       string
	SkipIfRunning bool
	Locks         []string
	Disabled      bool

	Raw string
}

func ParseCronLine(name, line string) (*CronType, error) {
	parts := strings.Fields(line)
	if len(parts) < 6 {
		return nil, fmt.Errorf("cron string seems invalid")
	}

	cron := &CronType{
		Name:       name,
		Raw:        line,
		Minute:     parts[0],
		Hour:       parts[1],
		DayOfMonth: parts[2],
		Month:      parts[3],
		DayOfWeek:  parts[4],
	}

	parts = parts[5:]

	for len(parts) > 0 && strings.Contains(parts[0], ":") {
		keyval := parts[0]
		parts = parts[1:]

		keyvalParts := strings.SplitN(keyval, ":", 2)
		key := keyvalParts[0]
		value := keyvalParts[1]

		switch strings.ToLower(key) {
		case "queue":
			cron.Queue = value
		case "skipifrunning":
			cron.SkipIfRunning = (value == "true" || value == "1")
		case "locks":
			cron.Locks = strings.Split(value, ",")
		default:
			//nothing yet!
		}
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("cron string seems invalid")
	}
	if cron.Queue == "" {
		return nil, fmt.Errorf("cron without queue is invalid")
	}

	cron.Command = strings.Join(parts, " ")
	return cron, nil
}

func (cron *CronType) MatchTime(t time.Time) bool {
	t = t.UTC()

	if !cronTimeCompare(cron.Minute, t.Minute()) {
		return false
	}

	if !cronTimeCompare(cron.Hour, t.Hour()) {
		return false
	}

	if !cronTimeCompare(cron.DayOfMonth, t.Day()) {
		return false
	}

	if !cronTimeCompare(cron.Month, int(t.Month())) {
		return false
	}

	if !cronTimeCompare(cron.DayOfWeek, int(t.Weekday())) {
		return false
	}

	return true
}

func cronTimeCompare(cronstr string, timeval int) bool {
	if cronstr == "*" {
		return true
	}

	for _, cronval := range strings.Split(cronstr, ",") {
		if strings.Contains(cronval, "-") {
			cronValParts := strings.SplitN(cronval, "-", 2)
			start, _ := strconv.Atoi(cronValParts[0])
			end, _ := strconv.Atoi(cronValParts[1])
			if timeval >= start && timeval <= end {
				return true
			}
		} else if strings.HasPrefix(cronval, "*/") && len(cronval) > 2 {
			divisor, _ := strconv.Atoi(cronval[2:])
			if timeval%divisor == 0 {
				return true
			}
		} else {
			cronval, _ := strconv.Atoi(cronval)
			if cronval == timeval {
				return true
			}
		}
	}

	return false
}

func (cron *CronType) Task() *tasklib.Task {
	return &tasklib.Task{
		Command: cron.Command,
		Cron:    cron.Name,
		Locks:   cron.Locks,
	}
}
