package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func SleepUntilNextMinute() {
	now := time.Now().Unix()
	last_minute := now - now%60
	next_minute := last_minute + 60
	sleep_for := next_minute - now

	time.Sleep(time.Duration(sleep_for) * time.Second)
}

func Md5sum(data interface{}) string {
	hasher := md5.New()

	switch data := data.(type) {
	case string:
		io.WriteString(hasher, data)
	case []byte:
		hasher.Write(data)
	default:
		panic("Invalid type passed to md5sum()")
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func ProcessExists(pid int) bool {
	_, err := os.FindProcess(int(pid))
	if err != nil {
		return false
	}
	return true
}

func HumanDuration(d time.Duration, fields int) string {
	secondDurations := []int64{31540000, 2628000, 86400, 3600, 60, 1}
	humanDurations := []string{"year", "month", "day", "hour", "minute", "second"}

	parts := []string{}
	seconds := int64(d.Seconds())
	if seconds == 0 {
		return "less than 1 second"
	}

	for i, sDuration := range secondDurations {
		if seconds >= sDuration {
			multiple := seconds / sDuration

			trail := "s"
			if multiple == 1 {
				trail = ""
			}
			parts = append(parts, fmt.Sprintf("%d %s%s", multiple, humanDurations[i], trail))
			seconds -= multiple * sDuration
		}
	}

	if len(parts) > fields {
		parts = parts[:fields]
	}

	return strings.Join(parts, " ")
}
