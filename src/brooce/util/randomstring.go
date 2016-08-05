package util

import (
	"math/rand"
	"sync"
	"time"
)

var randSource = rand.NewSource(time.Now().UnixNano())
var randSourceMutex = sync.Mutex{}

func RandomString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	const letterIdxBits = 6
	const letterIdxMask = 1<<letterIdxBits - 1
	const letterIdxMax = 63 / letterIdxBits

	b := make([]byte, length)

	for i, cache, remain := length-1, int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func int63() int64 {
	randSourceMutex.Lock()
	defer randSourceMutex.Unlock()
	return randSource.Int63()
}
