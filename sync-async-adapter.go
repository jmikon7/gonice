package gonice

import (
	"context"
	"errors"
	"sync"
	"time"
)

var TimeoutError = errors.New("timeout")
var AlreadyDestroyedError = errors.New("already destroyed")
var DuplicateInflightRequestError = errors.New("duplicate inflight request")

type AsyncService interface {
	SendRequest(interface{})
	GetResponseCh() chan interface{}
	ExtractIdFromRequest(interface{}) string
	ExtractIdFromResponse(interface{}) string
}

type UnrecognizedResponseCb func(string, interface{})

type SyncAsyncAdapter struct {
	requestContexts map[string]*RequestContext
	lock            sync.RWMutex
	service         AsyncService
	destroyed       bool
	stopListener    context.CancelFunc
	pool            *sync.Pool
}

func (s *SyncAsyncAdapter) Destroy() {
	s.stopListener()
}

func NewSyncAsyncAdapter(service AsyncService) *SyncAsyncAdapter {
	ctx, stopListener := context.WithCancel(context.Background())
	s := &SyncAsyncAdapter{
		requestContexts: make(map[string]*RequestContext),
		lock:            sync.RWMutex{},
		stopListener:    stopListener,
		destroyed:       false,
		pool: &sync.Pool{
			New: func() interface{} {
				return newRequestContext()
			},
		},
		service: service}
	go s.listen(ctx)
	return s
}

func (s *SyncAsyncAdapter) listen(ctx context.Context) {
	run := true
	for run {
		select {
		case response := <-s.service.GetResponseCh():
			s.applyResponse(response)
		case <-ctx.Done():
			run = false
		}
	}
	s.destroyed = true
}

func (s *SyncAsyncAdapter) applyResponse(response interface{}) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	requestId := s.service.ExtractIdFromResponse(response)
	if requestContext, ex := s.requestContexts[requestId]; ex {
		requestContext.listenCh <- response
	}
}

func (s *SyncAsyncAdapter) register(request interface{}) (*RequestContext, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	requestId := s.service.ExtractIdFromRequest(request)
	if _, ex := s.requestContexts[requestId]; ex {
		return nil, "", DuplicateInflightRequestError
	}
	requestContext := s.pool.Get().(*RequestContext)
	requestContext.listen()
	s.requestContexts[requestId] = requestContext
	return requestContext, requestId, nil
}

func (s *SyncAsyncAdapter) unregister(requestContext *RequestContext, requestId string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if requestContext, ex := s.requestContexts[requestId]; ex {
		delete(s.requestContexts, requestId)
		requestContext.cancel()
		s.pool.Put(requestContext)
	}
}

func (s *SyncAsyncAdapter) Do(request interface{}, waitTime time.Duration) (interface{}, error) {
	if s.destroyed {
		return nil, AlreadyDestroyedError
	}
	requestContext, requestId, err := s.register(request)
	if err != nil {
		return nil, err
	}
	timeout := time.NewTicker(waitTime)
	defer func() {
		timeout.Stop()
		s.unregister(requestContext, requestId)
	}()
	s.service.SendRequest(request)
	select {
	case response := <-requestContext.responseCh:
		return response, nil
	case <-timeout.C:
		return nil, TimeoutError
	}
}

func newRequestContext() *RequestContext {
	rc := &RequestContext{
		responseCh: make(chan interface{}),
		listenCh:   make(chan interface{})}
	return rc
}

type RequestContext struct {
	responseCh     chan interface{}
	listenCh       chan interface{}
	stop           context.CancelFunc
	ctx            context.Context
	cancelComplete *sync.WaitGroup
}

func (rc *RequestContext) cancel() {
	rc.stop()
	rc.cancelComplete.Wait()
}

func (rc *RequestContext) listen() {
	rc.ctx, rc.stop = context.WithCancel(context.Background())
	rc.cancelComplete = NewSingleWaitGroup()
	go func() {
		select {
		case buf := <-rc.listenCh:
			select {
			case rc.responseCh <- buf:
			case <-rc.ctx.Done():
			}
		case <-rc.ctx.Done():
		}
		rc.cancelComplete.Done()
	}()
}
