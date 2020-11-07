package gonice

import (
	"sync"
	"testing"
	"time"
)

func TestDispatch_ProcLessThenThreads(t *testing.T) {
	processors := 5
	threads := 100
	complete := NewWaitGroup(processors)
	processorsList := make([]Processor, processors)
	for i := 0; i < processors; i++ {
		processorsList[i] = newTestProcessor(complete)
	}
	Dispatch("data", processorsList, threads)

	if WaitTimeout(complete, time.Second) {
		t.Fatal("Timeout")
	}
}

func TestDispatch_ProcMoreThenThreads(t *testing.T) {
	processors := 100
	threads := 5
	complete := NewWaitGroup(processors)
	processorsList := make([]Processor, processors)
	for i := 0; i < processors; i++ {
		processorsList[i] = newTestProcessor(complete)
	}
	Dispatch("data", processorsList, threads)

	if WaitTimeout(complete, time.Second) {
		t.Fatal("Timeout")
	}
}

func newTestProcessor(complete *sync.WaitGroup) *TestProcessor {
	return &TestProcessor{
		complete: complete}
}

type TestProcessor struct {
	complete *sync.WaitGroup
}

func (p *TestProcessor) Do(v interface{}) {
	p.complete.Done()
}
