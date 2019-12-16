package titan

import (
	"os"
	"strings"
)

type Health struct {
	Status string `json:"status"`
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
	subject string
}

func (h *DefaultHandlers) Routes(r Router) {
	basePath := "/" + strings.Join(strings.Split(h.subject, "."), "/")
	r.RegisterJson("GET", basePath+"/health", h.Health)
	r.RegisterJson("GET", basePath+"/info", h.AppInfo)
}

func (h *DefaultHandlers) Health(ctx *Context) (*Health, error) {
	return &Health{Status: "UP"}, nil
}

//see BuildInfoSource.java
func (h *DefaultHandlers) AppInfo(ctx *Context) (*AppInfo, error) {
	return &AppInfo{Build: BuildInfo{
		Version: os.Getenv("BUILD_VERSION"),
		Date:    os.Getenv("BUILD_DATE"),
		Tag:     os.Getenv("BUILD_TAG"),
	}}, nil
}
