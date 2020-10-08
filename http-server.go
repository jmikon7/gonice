package gonice

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Endpoint interface {
	Register(r *mux.Router)
}

type Server struct {
	addr   string
	server *http.Server
	router *mux.Router
}

func Create(host string, port int) *Server {
	return &Server{
		addr:   fmt.Sprintf("%s:%d", host, port),
		router: mux.NewRouter(),
	}
}

func (s *Server) Start() *Server {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return s
}

func (s *Server) WithEndpoints(endpoints ...Endpoint) *Server {
	if endpoints != nil {
		for _, ep := range endpoints {
			ep.Register(s.router)
		}
	}
	return s
}

func (s *Server) Shutdown(timeout time.Duration) {
	if s.server == nil {
		return
	}
	if timeout == 0 {
		timeout = time.Second * 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		println(err.Error())
	}
}
