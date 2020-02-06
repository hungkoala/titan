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

	WaitOrTimeout(s.t, s.Server.Started, "Server start timed out")
}

type TestWaitGroup struct {
	*sync.WaitGroup
	t    *testing.T
	done chan interface{}
}

func NewTestWaitGroup(t *testing.T) *TestWaitGroup {
	wg := sync.WaitGroup{}
	done := make(chan interface{}, 1)
	return &TestWaitGroup{&wg, t, done}
}

func (wg *TestWaitGroup) WaitOrTimeout() {
	wg.WaitOrTimeoutFor(1)
}

func (wg *TestWaitGroup) WaitOrTimeoutFor(timeout int) {
	wg.WaitOrTimeoutForWithMessage(timeout, "WaitGroup timed out")
}

func (wg *TestWaitGroup) WaitOrTimeoutForWithMessage(timeout int, timeoutMessage string) {
	go func() {
		wg.WaitGroup.Wait()
		close(wg.done)
	}()

	WaitOrTimeoutFor(wg.t, wg.done, timeout, timeoutMessage)
}

func WaitOrTimeout(t *testing.T, ch chan interface{}, msg string) {
	WaitOrTimeoutFor(t, ch, 1, msg)
}

func WaitOrTimeoutFor(t *testing.T, ch chan interface{}, timeout int, msg string) {
	select {
	case <-ch:
		break
	case <-time.After(time.Duration(timeout) * time.Second):
		if t != nil {
			t.Logf("%s after %d seconds", msg, timeout)
			t.Fail()
		} else {
			panic(fmt.Sprintf("%s after %d seconds", msg, timeout))
		}
	}
}

type TestServers struct {
	servers []*titan.Server
}

func NewTestServers(servers []*titan.Server) TestServers {
	return TestServers{servers}
}

func (s *TestServers) Start() {
	wg := NewTestWaitGroup(nil)

	for _, server := range s.servers {
		wg.Add(1)
		go func(server *titan.Server) {
			server.Start()
		}(server)
		go func(server *titan.Server, wg *TestWaitGroup) {
			<-server.Started
			wg.Done()
		}(server, wg)
	}

	wg.WaitOrTimeoutForWithMessage(len(s.servers), "Servers start timed out")
}

func (s *TestServers) Stop() {
	for _, server := range s.servers {
		server.Stop()
	}
}
