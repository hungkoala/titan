package titan

import (
	"os"
	"runtime"
	"strings"
)

const (
	HEALTH_CHECK       = "health_check"
	HEALTH_CHECK_REPLY = "health_check_reply"
	UP                 = "UP"

	MONITORING_CHECK       = "monitoring_check"
	MONITORING_CHECK_REPLY = "monitoring_check_reply"
)

type Health struct {
	Status   string `json:"status"`
	HostName string `json:"hostName"`
	Subject  string `json:"subject"`
	Language string `json:"language"`
}

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

type AppInfo struct {
	Build BuildInfo `json:"build"`
}

type BuildInfo struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Tag     string `json:"tag"`
}

type DefaultHandlers struct {
	Subject string
}

func (h *DefaultHandlers) Register(r Router) {
	basePath := ""
	if h.Subject != "" {
		basePath = "/" + strings.Join(strings.Split(h.Subject, "."), "/")
	}

	r.RegisterJson("GET", basePath+"/health", h.Health)
	r.RegisterJson("GET", basePath+"/info", h.AppInfo)
}

func (h *DefaultHandlers) Health(ctx *Context) (*Health, error) {
	health := h.healthCheck()
	return &health, nil
}

//see BuildInfoSource.java
func (h *DefaultHandlers) AppInfo(ctx *Context) (*AppInfo, error) {
	return &AppInfo{Build: BuildInfo{
		Version: os.Getenv("BUILD_VERSION"),
		Date:    os.Getenv("BUILD_DATE"),
		Tag:     os.Getenv("BUILD_TAG"),
	}}, nil
}

func (h *DefaultHandlers) Subscribe(s *MessageSubscriber) {
	s.Register(HEALTH_CHECK, "", func(p *Message) error {
		return GetDefaultClient().Publish(NewBackgroundContext(), HEALTH_CHECK_REPLY, h.healthCheck())
	})
	s.Register(MONITORING_CHECK, "", func(p *Message) error {
		return GetDefaultClient().Publish(NewBackgroundContext(), MONITORING_CHECK_REPLY, h.monitoringCheck())
	})
}

func (h *DefaultHandlers) healthCheck() Health {
	name, _ := os.Hostname()
	health := Health{
		Status:   UP,
		HostName: name,
		Subject:  h.Subject,
		Language: "Go",
	}
	return health
}
func (h *DefaultHandlers) monitoringCheck() Monitoring {
	name, _ := os.Hostname()
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	process, err := GetCpuUsage()

	monitoring := Monitoring{
		Status:   UP,
		HostName: name,
		Subject:  h.Subject,
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
	} else {
		//logger.Warn(fmt.Sprintf("monitoring check error %+v", err))
	}
	return monitoring
}
