package main

import (
	"os/exec"
	"sync"
	"time"
)

var statusLock = sync.Mutex{}
var workingThreads = map[int]bool{}
var lastStatusChangeTime = time.Now()

func announceStatusWorking(threadid int) {
	statusLock.Lock()
	defer statusLock.Unlock()

	workingThreads[threadid] = true
	lastStatusChangeTime = time.Now()
}

func announceStatusWaiting(threadid int) {
	statusLock.Lock()
	defer statusLock.Unlock()

	delete(workingThreads, threadid)
	lastStatusChangeTime = time.Now()
}

func suicider() {
	if !config.Suicide.Enabled {
		return
	}

	for {
		time.Sleep(time.Minute)
		checkSuicide()
	}
}

func checkSuicide() {
	statusLock.Lock()
	defer statusLock.Unlock()

	if !config.Suicide.Enabled {
		return
	}

	statusChangeAgo := time.Since(lastStatusChangeTime)
	countWorking := len(workingThreads)

	if countWorking > 0 || int(statusChangeAgo.Seconds()) < config.Suicide.Time {
		return
	}

	logger.Printf("Suicide threshold reached! %v working threads; last event %v ago! Running: %v",
		countWorking, statusChangeAgo, config.Suicide.Command)
	exec.Command("bash", "-c", config.Suicide.Command).Run()
}
