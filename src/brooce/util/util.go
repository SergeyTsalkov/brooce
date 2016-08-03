package util

import "time"

func SleepUntilNextMinute() {
	now := time.Now().Unix()
	last_minute := now - now%60
	next_minute := last_minute + 60
	sleep_for := next_minute - now

	time.Sleep(time.Duration(sleep_for) * time.Second)
}
