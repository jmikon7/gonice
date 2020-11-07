package gonice

import (
	"bytes"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

var HelloBytes = []byte("Hello World")

func TestHttpService_HelloWorld(t *testing.T) {
	s := Create("", 8080).WithEndpoints(new(Foo)).Start()
	defer s.Shutdown(0)
	if !waitForSocketAvalable("localhost:8080", time.Minute) {
		t.Fatal("Timeout waiting socket to connect")
	}
	resp, err := http.Get("http://localhost:8080/foo")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(HelloBytes, body) != 0 {
		t.Fatal("Wrong answer")
	}

}

type Foo int

func (e *Foo) Register(r *mux.Router) {
	r.HandleFunc("/foo", e.accept())
}
func (e *Foo) accept() func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(HelloBytes); err != nil {
			println(err.Error())
		}
	}
}

func waitForSocketAvalable(connect string, waitTime time.Duration) bool {
	socket, err := net.DialTimeout("tcp", connect, waitTime)
	defer socket.Close()
	return err == nil
}
