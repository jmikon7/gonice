package gonice

import "sync"

type Processor interface {
	Do(input interface{})
}

func Dispatch(inputData interface{}, toProcessors []Processor, inThreads int) {
	if len(toProcessors) == 0 {
		return
	}
	ch := make(chan Processor)
	wg := new(sync.WaitGroup)
	wg.Add(len(toProcessors))
	for i := 0; i < inThreads; i++ {
		go func(input interface{}) {
			for processor := range ch {
				processor.Do(input)
				wg.Done()
			}
		}(inputData)
	}
	for _, processor := range toProcessors {
		ch <- processor
	}
	wg.Wait()
	close(ch)
}
