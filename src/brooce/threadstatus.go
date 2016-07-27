package main

import (
	"log"
	"os/exec"
	"sync"
	"time"

	"brooce/config"
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
	if !config.Config.Suicide.Enabled {
		return
	}

	log.Println("After", config.Config.Suicide.Time, "seconds of inactivity, we will run:", config.Config.Suicide.Command)

	for {
		time.Sleep(time.Minute)
		checkSuicide()
	}
}

func checkSuicide() {
	statusLock.Lock()
	defer statusLock.Unlock()

	if !config.Config.Suicide.Enabled {
		return
	}

	statusChangeAgo := time.Since(lastStatusChangeTime)
	countWorking := len(workingThreads)

	if countWorking > 0 || int(statusChangeAgo.Seconds()) < config.Config.Suicide.Time {
		return
	}

	logger.Printf("Suicide threshold reached! %v working threads; last event %v ago! Running: %v",
		countWorking, statusChangeAgo, config.Config.Suicide.Command)
	exec.Command("bash", "-c", config.Config.Suicide.Command).Run()
}
