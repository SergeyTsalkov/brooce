package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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

func ProcessExists(pid int) bool {
	return FileExists(fmt.Sprintf("/proc/%v", pid))
}
