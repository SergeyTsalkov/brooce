package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func SleepUntilNextInterval(interval int) {
	now := time.Now().Unix()
	last_interval := now - now%int64(interval)
	next_interval := last_interval + int64(interval)
	sleep_for := next_interval - now

	time.Sleep(time.Duration(sleep_for) * time.Second)
}

func SleepUntilNextMinute() {
	SleepUntilNextInterval(60)
}

func Md5sum(data interface{}) string {
	hasher := md5.New()

	switch data := data.(type) {
	case string:
		io.WriteString(hasher, data)
	case []byte:
		hasher.Write(data)
	default:
		log.Fatalln("Invalid type passed to md5sum()")
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
	// TODO: better cross-system process exists handling
	if runtime.GOOS != "darwin" {
		return FileExists(fmt.Sprintf("/proc/%v", pid))
	}

	_, err := os.FindProcess(pid)
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
