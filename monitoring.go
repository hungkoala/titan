package titan

import (
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"

	"logur.dev/logur"
)

const (
	MONITORING_CHECK       = "monitoring_check"
	MONITORING_CHECK_REPLY = "monitoring_check_reply"
)

var msgInNum uint64
var msgOutNum uint64
var msgErrNum uint64

var _monitoringSubscription *nats.Subscription

type Monitoring struct {
	Status       string `json:"status"`
	HostName     string `json:"hostName"`
	Subject      string `json:"subject"`
	Alloc        uint64 `json:"alloc"`        // currently allocated number of bytes on the heap
	TotalAlloc   uint64 `json:"totalAlloc"`   //cumulative max bytes allocated on the heap (will not decrease),
	Sys          uint64 `json:"sys"`          //total memory obtained from the OS
	Mallocs      uint64 `json:"mallocs"`      //number of allocations
	Frees        uint64 `json:"frees"`        //number  deallocations
	LiveObjects  uint64 `json:"liveObjects"`  //live objects (mallocs - frees)
	PauseTotalNs uint64 `json:"pauseTotalNs"` //total GC pauses since the app has started,
	NumGC        uint32 `json:"numGC"`        // number of completed GC cycles

	NumGoroutine int `json:"numGoroutine"`

	Pid int     `json:"pid"` // process id
	Cpu float64 `json:"cpu"` // cpu usage

	MsgInNum  uint64 `json:"msgInNum"`  // number of  in  messages
	MsgOutNum uint64 `json:"msgOutNum"` // number of  out  messages
	MsgErrNum uint64 `json:"msgErrNum"` // total of errors

	MsgPendingNum int `json:"msgPendingNum"` //pending message in queue

	Language    string `json:"language"`
	RequestTime int64  `json:"requestTime"`
}

func DoMonitoringCheck(subject string, m *Message) Monitoring {
	name, _ := os.Hostname()
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	process, err := GetCpuUsage()

	var pendingMsg int
	if _monitoringSubscription != nil && _monitoringSubscription.IsValid() {
		pMsgs, _, err := _monitoringSubscription.Pending()
		if err != nil {
			pendingMsg = pMsgs
		}
	}

	monitoring := Monitoring{
		Status:   UP,
		HostName: name,
		Subject:  subject,
		// Misc memory stats
		Alloc:      rtm.Alloc,
		TotalAlloc: rtm.TotalAlloc,
		Sys:        rtm.Sys,
		Mallocs:    rtm.Mallocs,
		Frees:      rtm.Frees,

		// Live objects = Mallocs - Frees
		LiveObjects: rtm.Mallocs - rtm.Frees,

		// GC Stats
		PauseTotalNs:  rtm.PauseTotalNs,
		NumGC:         rtm.NumGC,
		NumGoroutine:  runtime.NumGoroutine(),
		Language:      "Go",
		MsgInNum:      atomic.LoadUint64(&msgInNum),
		MsgOutNum:     atomic.LoadUint64(&msgOutNum),
		MsgErrNum:     atomic.LoadUint64(&msgErrNum),
		RequestTime:   time.Now().UnixNano() / int64(time.Millisecond),
		MsgPendingNum: pendingMsg,
	}

	if err == nil {
		monitoring.Pid = process.pid
		monitoring.Cpu = process.cpu
	}

	return monitoring
}

func BeginRequest(logger logur.Logger, request RequestInterface) {
	go func(logger logur.Logger, h http.Header) {
		if r := recover(); r != nil {
			logger.Info("BeginRequest panic recovered")
		}

		MsgInNumAdd(1)

		// log  too high latency
		requestTimeStr := h.Get(XRequestTime)

		if requestTimeStr != "" {
			requestTime, err := strconv.ParseInt(requestTimeStr, 10, 64)
			if err == nil {
				duration := (time.Now().UnixNano() - requestTime) / 1000000
				if duration > 2000 { // 1 second
					logger.Warn("request latency is too high", map[string]interface{}{"time": duration})
				}
			}
		} else {
			logger.Warn("Go Request time not found")
		}
	}(logger, request.GetHeaders().Clone())
}

func EndRequest(logger logur.Logger, response ResponseInterface) {
	go func(logger logur.Logger, response ResponseInterface) {
		if r := recover(); r != nil {
			logger.Info("EndRequest panic recovered")
		}
		MsgOutNumAdd(1)
		if response.GetStatusCode() < 200 || response.GetStatusCode() >= 300 {
			MsgErrNumAdd(1)
		}
	}(logger, response)
}

func MsgInNumAdd(v uint64) {
	if msgInNum >= 18446744073709551000 {
		atomic.StoreUint64(&msgInNum, 0)
	}
	atomic.AddUint64(&msgInNum, v)
}

func MsgOutNumAdd(v uint64) {
	if msgOutNum >= 18446744073709551000 {
		atomic.StoreUint64(&msgOutNum, 0)
	}
	atomic.AddUint64(&msgOutNum, v)
}

func MsgErrNumAdd(v uint64) {
	if msgErrNum >= 18446744073709551000 {
		atomic.StoreUint64(&msgErrNum, 0)
	}
	atomic.AddUint64(&msgErrNum, v)
}

func MsgCountLoad() uint64 {
	return atomic.LoadUint64(&msgInNum) - atomic.LoadUint64(&msgOutNum)
}
