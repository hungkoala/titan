package titan

import (
	"os"
	"strings"
)

const (
	HEALTH_CHECK       = "health_check"
	HEALTH_CHECK_REPLY = "health_check_reply"
	UP                 = "UP"
)

type Health struct {
	Status   string `json:"status"`
	HostName string `json:"hostName"`
	Subject  string `json:"subject"`
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
	name, _ := os.Hostname()
	return &Health{Status: UP, HostName: name, Subject: h.Subject}, nil
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
		name, _ := os.Hostname()
		return GetDefaultClient().Publish(NewBackgroundContext(), HEALTH_CHECK_REPLY, Health{
			Status:   UP,
			HostName: name,
			Subject:  h.Subject,
		})
	})
}
