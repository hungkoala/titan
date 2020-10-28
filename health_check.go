package titan

import (
	"os"
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
	Language string `json:"language"`
}

func HealthCheck(ctx *Context, subject string) (*Health, error) {
	client := GetDefaultClient()
	req, _ := NewReqBuilder().
		Get(SubjectToUrl(subject, "health")).
		Subject(subject).
		Build()
	var resp = &Health{}
	err := client.SendAndReceiveJson(ctx, req, &resp)
	return resp, err
}

func (h *DefaultHandlers) DoHealthCheck() Health {
	name, _ := os.Hostname()
	health := Health{
		Status:   UP,
		HostName: name,
		Subject:  h.Subject,
		Language: "Go",
	}
	return health
}
