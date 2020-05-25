package test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"gitlab.com/silenteer-oss/titan"
)

type TestServer struct {
	Server titan.IServer
	t      *testing.T
}

func NewTestServer(t *testing.T, server titan.IServer) *TestServer {
	return &TestServer{server, t}
}

func (s *TestServer) Start() {
	ch := make(chan interface{}, 1)
	go func() {
		s.Server.Start(ch)
	}()

	WaitOrTimeout(s.t, ch, "Server start timed out")
}

func (s *TestServer) Stop() {
	s.Server.Stop()
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
	servers []titan.IServer
}

func NewTestServers(servers []titan.IServer) TestServers {
	return TestServers{servers}
}

func (s *TestServers) Start() {
	wg := NewTestWaitGroup(nil)

	for _, server := range s.servers {
		wg.Add(1)
		go func(server titan.IServer, wg *TestWaitGroup) {
			ch := make(chan interface{}, 1)
			go func(s titan.IServer) {
				s.Start(ch)
			}(server)
			select {
			case <-ch:
				wg.Done()
			case <-time.After(time.Duration(20) * time.Second):
				panic(fmt.Sprintf("Servers start timed out after %d seconds", 1))
			}
		}(server, wg)
	}

	wg.WaitOrTimeoutForWithMessage(len(s.servers)*20, "Servers start timed out")
}

func (s *TestServers) Stop() {
	for _, server := range s.servers {
		server.Stop()
	}
}
