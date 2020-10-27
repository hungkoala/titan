package titan

import (
	"os"
	"runtime"
)

const (
	MONITORING_CHECK       = "monitoring_check"
	MONITORING_CHECK_REPLY = "monitoring_check_reply"
)

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
	Language string  `json:"language"`
}

func DoMonitoringCheck(subject string) Monitoring {
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
	}

	if err == nil {
		monitoring.Pid = process.pid
		monitoring.Cpu = process.cpu
	}
	return monitoring
}
