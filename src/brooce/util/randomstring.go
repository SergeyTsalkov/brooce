package util

import (
	"math/rand"
	"sync"
	"time"
)

var myrand = rand.New(rand.NewSource(time.Now().UnixNano()))
var myrandMutex = sync.Mutex{}

func RandomString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	output := make([]byte, length)
	for i := 0; i < length; i++ {
		output[i] = letterBytes[int31n(len(letterBytes))]
	}

	return string(output)
}

func int31n(n int) int32 {
	myrandMutex.Lock()
	defer myrandMutex.Unlock()
	return myrand.Int31n(int32(n))
}
