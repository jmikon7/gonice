package gonice

import (
	"context"
	"time"
)

func TickerTask(onTick func(), onComplete func(), every time.Duration) context.CancelFunc {
	if onTick == nil {
		return nil
	}
	tc := time.NewTicker(every)
	ctx, f := context.WithCancel(context.Background())
	go func() {
		run := true
		for run {
			select {
			case <-tc.C:
				onTick()
			case <-ctx.Done():
				run = false
			}
		}
		tc.Stop()
		if onComplete != nil {
			onComplete()
		}
	}()
	return f
}
