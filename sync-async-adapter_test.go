package gonice

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const TestTimeout = time.Millisecond * 50
const ZeroProcessTime = 0

type TestFunc func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T)
type PathType int

const (
	Happy   PathType = 1
	Fail    PathType = 2
	Timeout PathType = 3
)

func BenchmarkSyncAsync_Do(b *testing.B) {
	service := newTestAsyncService(ZeroProcessTime)
	adapter := NewSyncAsyncAdapter(service)
	defer adapter.Destroy()

	request := "request"
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := adapter.Do(request, time.Second)
		if err != nil {
			b.Fatalf("%s %v", request, err)

		}
	}
}

func TestSyncAsync_MultiZero(t *testing.T) {
	threads := 500
	count := 100
	initTest(ZeroProcessTime, threads, t, func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T) {
		test(thread, count, adapter, wg, t, time.Nanosecond, Fail)
	})
}

func TestSyncAsync_SingleZero(t *testing.T) {
	threads := 1
	count := 500000
	initTest(ZeroProcessTime, threads, t, func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T) {
		test(thread, count, adapter, wg, t, time.Nanosecond, Fail)
	})
}

func TestSyncAsync_MultiThreadTimeout(t *testing.T) {
	threads := 5
	count := 20
	initTest(time.Second, threads, t, func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T) {
		test(thread, count, adapter, wg, t, time.Millisecond, Timeout)
	})
}

func TestSyncAsync_SingleThreadTimeout(t *testing.T) {
	threads := 1
	count := 100
	initTest(time.Second, threads, t, func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T) {
		test(thread, count, adapter, wg, t, time.Millisecond, Timeout)
	})
}

func TestSyncAsync_MultiThreadHappy(t *testing.T) {
	threads := 500
	count := 100
	initTest(ZeroProcessTime, threads, t, func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T) {
		test(thread, count, adapter, wg, t, TestTimeout, Happy)
	})
}

func TestSyncAsync_SingleThreadHappy(t *testing.T) {
	threads := 1
	count := 500000
	initTest(ZeroProcessTime, threads, t, func(adapter *SyncAsyncAdapter, wg *sync.WaitGroup, thread int, t *testing.T) {
		test(thread, count, adapter, wg, t, TestTimeout, Happy)
	})
}

func initTest(processTime time.Duration, threads int, t *testing.T, testFunc TestFunc) {
	service := newTestAsyncService(processTime)
	adapter := NewSyncAsyncAdapter(service)
	defer adapter.Destroy()
	wg := NewWaitGroup(threads)
	for th := 0; th < threads; th++ {
		go testFunc(adapter, wg, th, t)
	}
	wg.Wait()

}

func test(thread int, count int, adapter *SyncAsyncAdapter, wg *sync.WaitGroup, t *testing.T, timeout time.Duration, path PathType) {
	defer func() {
		recover()
		wg.Done()
	}()
	var prevErr error
	for i := 0; i < count; i++ {
		request := fmt.Sprintf("request: %d %d", thread, i)
		response, err := adapter.Do(request, timeout)
		if path == Happy {
			if err != nil {
				t.Fatalf("%s %v", request, err)
			}
			if request != response {
				t.Fatalf("Not the same in={%s} out={%s}  %v", request, response, prevErr)
			}
		} else if path == Fail {
			if err == nil {
				if request != response {
					t.Fatalf("Not the same in={%s} out={%s}  %v", request, response, prevErr)
				}
			}
		} else if path == Timeout {
			if err != TimeoutError {
				t.Fatal("Must be timeout error")
			}
		}
		prevErr = err
	}
}

func newTestAsyncService(processTime time.Duration) *TestAsyncService {
	s := &TestAsyncService{
		in:          make(chan interface{}),
		out:         make(chan interface{}),
		processTime: processTime}
	for i := 0; i < 10; i++ {
		go s.listen(i)
	}
	return s
}

type TestAsyncService struct {
	in          chan interface{}
	out         chan interface{}
	processTime time.Duration
}

func (s *TestAsyncService) listen(th int) {
	for value := range s.in {
		if s.processTime != 0 {
			time.Sleep(s.processTime)
		}
		s.out <- value
	}
}
func (s *TestAsyncService) ExtractIdFromRequest(request interface{}) string {
	return request.(string)
}
func (s *TestAsyncService) ExtractIdFromResponse(response interface{}) string {
	return response.(string)
}
func (s *TestAsyncService) SendRequest(request interface{}) {
	s.in <- request
}
func (s *TestAsyncService) GetResponseCh() chan interface{} {
	return s.out
}
