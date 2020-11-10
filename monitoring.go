package titan

import (
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"logur.dev/logur"
)

const (
	MONITORING_CHECK       = "monitoring_check"
	MONITORING_CHECK_REPLY = "monitoring_check_reply"
)

var msgNum int64
var msgTotal uint64

type Monitoring struct {
	Status       string `json:"status"`
	HostName     string `json:"hostName"`
	Subject      string `json:"subject"`
	Alloc        uint64 // currently allocated number of bytes on the heap
	TotalAlloc   uint64 //cumulative max bytes allocated on the heap (will not decrease),
	Sys          uint64 //total memory obtained from the OS
	Mallocs      uint64 //number of allocations
	Frees        uint64 //number  deallocations
	LiveObjects  uint64 //live objects (mallocs - frees)
	PauseTotalNs uint64 //total GC pauses since the app has started,
	NumGC        uint32 // number of completed GC cycles

	NumGoroutine int

	Pid      int     // process id
	Cpu      float64 // cpu usage
	MsgNum   int64   // number of processing messages
	MsgTotal uint64  // number of messages have been processed
	Language string  `json:"language"`
}

func DoMonitoringCheck(subject string, m *Message) Monitoring {
	name, _ := os.Hostname()
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	process, err := GetCpuUsage()

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
		PauseTotalNs: rtm.PauseTotalNs,
		NumGC:        rtm.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		Language:     "Go",
		MsgNum:       atomic.LoadInt64(&msgNum),
		MsgTotal:     atomic.LoadUint64(&msgTotal),
	}

	if err == nil {
		monitoring.Pid = process.pid
		monitoring.Cpu = process.cpu
	}

	return monitoring
}

func BeginRequest(logger logur.Logger, header http.Header) {
	go func(logger logur.Logger, h http.Header) {
		if r := recover(); r != nil {
			logger.Info("BeginRequest panic recovered")
		}
		MsgCountAdd(1)
		//18446744073709551000 is maximum of unint64
		if msgTotal >= 18446744073709551000 {
			atomic.StoreUint64(&msgTotal, 0)
		}
		atomic.AddUint64(&msgTotal, 1)

		// log  too high latency

		requestTimeStr := h.Get(XRequestTime)

		if requestTimeStr != "" {
			requestTime, err := strconv.ParseInt(requestTimeStr, 10, 64)
			if err == nil {
				tine := (time.Now().UnixNano() - requestTime) / 1000000
				if tine > 2000 { // 1 second
					logger.Warn("request latency is too high", map[string]interface{}{"time": tine})
				}
			}
		}
	}(logger, header.Clone())
}

func EndRequest(logger logur.Logger, header http.Header) {
	go func(logger logur.Logger, header http.Header) {
		if r := recover(); r != nil {
			logger.Info("EndRequest panic recovered")
		}
		MsgCountAdd(-1)
	}(logger, header)
}

func MsgCountAdd(v int64) {
	atomic.AddInt64(&msgNum, v)
}

func MsgCountLoad() int64 {
	return atomic.LoadInt64(&msgNum)
}
