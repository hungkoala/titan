package titan

import (
	"fmt"
	"os"
	"strings"
)

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
	health := h.DoHealthCheck()
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
	healthCheckSubject := fmt.Sprintf("%s_%s", HEALTH_CHECK, strings.ReplaceAll(hostname, " ", "_"))
	s.Register(healthCheckSubject, "", func(p *Message) error {
		return GetDefaultClient().Publish(NewBackgroundContext(), HEALTH_CHECK_REPLY, h.DoHealthCheck())
	})
	s.Register(MONITORING_CHECK, "", func(m *Message) error {
		return GetDefaultClient().Publish(NewBackgroundContext(), MONITORING_CHECK_REPLY, DoMonitoringCheck(h.Subject, m))
	})
}
