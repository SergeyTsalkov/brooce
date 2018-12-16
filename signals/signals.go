package signals

import (
  "os"
  "os/signal"
  "sync"
  "sync/atomic"
)

var wg sync.WaitGroup
var done = new(int64)

func Start() {
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  wg.Add(1)

  go func() {
    <-c
    signal.Stop(c)
    wg.Done()
    atomic.AddInt64(done, 1)
  }()
}

func WaitForShutdownRequest() {
  wg.Wait()
}

func WasShutdownRequested() bool {
  return atomic.LoadInt64(done) > 0
}
