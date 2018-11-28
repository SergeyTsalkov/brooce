package suicide

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"brooce/config"
)

var statusLock = sync.Mutex{}
var workingThreads = map[string]bool{}
var lastStatusChangeTime = time.Now()

func Start() {
	if !config.Config.Suicide.Enable {
		return
	}

	log.Println("After", config.Config.Suicide.Time, "seconds of inactivity, we will run:", config.Config.Suicide.Command)

	go func() {
		for {
			time.Sleep(time.Minute)
			check()
		}
	}()
}

func ThreadIsWorking(threadName string) {
	statusLock.Lock()
	defer statusLock.Unlock()

	workingThreads[threadName] = true
	lastStatusChangeTime = time.Now()
}

func ThreadIsWaiting(threadName string) {
	statusLock.Lock()
	defer statusLock.Unlock()

	delete(workingThreads, threadName)
	lastStatusChangeTime = time.Now()
}

func check() {
	statusLock.Lock()
	defer statusLock.Unlock()

	if !config.Config.Suicide.Enable {
		return
	}

	statusChangeAgo := int(time.Since(lastStatusChangeTime).Seconds())
	countWorking := len(workingThreads)

	if countWorking > 0 || statusChangeAgo < config.Config.Suicide.Time {
		return
	}

	log.Printf("Suicide threshold reached! %v working threads; last event %v ago! Running: %v",
		countWorking, statusChangeAgo, config.Config.Suicide.Command)
	exec.Command("bash", "-c", config.Config.Suicide.Command).Run()
	os.Exit(0)
}
