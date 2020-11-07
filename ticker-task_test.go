package gonice

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTickerTask(t *testing.T) {
	var count int32 = 0
	var complete int32 = 0
	var N int32 = 5

	cancel := TickerTask(func() {
		atomic.AddInt32(&count, 1)
	}, func() {
		atomic.AddInt32(&complete, 1)
	}, time.Second)

	time.Sleep(time.Second*time.Duration(N) + time.Millisecond*500)
	ticks := atomic.LoadInt32(&count)
	if ticks != N {
		t.Fatalf("Must be %d ticks, but received only %d", N, ticks)
	}
	cancel()
	time.Sleep(time.Second)

	if atomic.LoadInt32(&complete) != 1 {
		t.Fatalf("OnComplete must be fired")
	}

}
