package gonice

import (
	"testing"
	"time"
)

func TestWaitTimeout_Happy(t *testing.T) {
	wg := NewSingleWaitGroup()
	go func() {
		time.Sleep(time.Second)
		wg.Done()
	}()
	if WaitTimeout(wg, time.Second*2) {
		t.Fatal("No timeout at this point")
	}
}

func TestWaitTimeout_Fail(t *testing.T) {
	wg := NewSingleWaitGroup()
	go func() {
		time.Sleep(time.Second * 2)
		wg.Done()
	}()
	if !WaitTimeout(wg, time.Second) {
		t.Fatal("Must be timeout at this point")
	}
}
