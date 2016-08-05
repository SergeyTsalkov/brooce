package cron

import (
	"errors"
	"strconv"
	"strings"
	"time"

	tasklib "brooce/task"
)

// Example cron entry:
// SET "brooce:cron:jobs:schedule_postprocess" "* * * * * queue:common gorunjob schedule_postprocess"

type cronType struct {
	minute     string
	hour       string
	dayOfMonth string
	month      string
	dayOfWeek  string

	queue   string
	command string
}

func parseCronLine(line string) (*cronType, error) {
	cron := &cronType{}

	parts := strings.Fields(line)
	if len(parts) < 6 {
		return nil, errors.New("cron string seems invalid")
	}

	cron.minute = parts[0]
	cron.hour = parts[1]
	cron.dayOfMonth = parts[2]
	cron.month = parts[3]
	cron.dayOfWeek = parts[4]

	parts = parts[5:]

	for len(parts) > 0 && strings.Contains(parts[0], ":") {
		keyval := parts[0]
		parts = parts[1:]

		keyvalParts := strings.SplitN(keyval, ":", 2)
		key := keyvalParts[0]
		value := keyvalParts[1]

		switch key {
		case "queue":
			cron.queue = value
		default:
			//nothing yet!
		}
	}

	if len(parts) == 0 {
		return nil, errors.New("cron string seems invalid")
	}

	cron.command = strings.Join(parts, " ")
	return cron, nil
}

func (cron *cronType) matchTime(t time.Time) bool {
	t = t.UTC()

	if !cronTimeCompare(cron.minute, t.Minute()) {
		return false
	}

	if !cronTimeCompare(cron.hour, t.Hour()) {
		return false
	}

	if !cronTimeCompare(cron.dayOfMonth, t.Day()) {
		return false
	}

	if !cronTimeCompare(cron.month, int(t.Month())) {
		return false
	}

	if !cronTimeCompare(cron.dayOfWeek, int(t.Weekday())) {
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

func (cron *cronType) task() *tasklib.Task {
	return &tasklib.Task{Command: cron.command}
}
