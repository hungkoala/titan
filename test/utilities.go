package test

import (
	"fmt"
	"gitlab.com/silenteer/titan"
	"sync"
	"testing"
	"time"
)

type TestServer struct {
	*titan.Server
	t *testing.T
}

func NewTestServer(t *testing.T, server *titan.Server) *TestServer {
	return &TestServer{server, t}
}

func (s *TestServer) Start() {

	go func() { s.Server.Start() }()

	WaitOrTimedOut(s.t, s.Server.Started, "Server start timed out")
}

type TestWaitGroup struct {
	*sync.WaitGroup
	t *testing.T
	done chan interface{}
}

func NewTestWaitGroup(t *testing.T) *TestWaitGroup {
	wg := sync.WaitGroup{}
	wg.Add(1)

	done := make(chan interface{}, 1)

	return &TestWaitGroup{&wg, t, done}
}

func (wg *TestWaitGroup) Wait() {
	go func() {
		wg.WaitGroup.Wait()
		close(wg.done)
	} ()

	WaitOrTimedOut(wg.t, wg.done, "WaitGroup timed out")
}

func WaitOrTimedOut(t *testing.T, ch chan interface{}, msg string) {
	select {
	case <- ch:
		break
	case <- time.After(1 * time.Second):
		fmt.Println(msg + " after 1 second")
		t.Fail()
	}
}

