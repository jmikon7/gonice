package gonice

import (
	"sync"
	"time"
)

func NewWaitGroup(n int) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(n)
	return wg
}

func NewSingleWaitGroup() *sync.WaitGroup {
	return NewWaitGroup(1)
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan interface{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
