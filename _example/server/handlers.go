package server

import (
	"encoding/json"

	"gitlab.com/silenteer/go-nats/nats"
)

type Handler struct {
	userService *UserService
}

func NewHandler(userService *UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Get(c *nats.Context, r *nats.Request) *nats.Response {
	c.Logger().Info("handler handles request")
	data := struct {
		RequestId   interface{}      `json:"RequestId"`
		QueryParams nats.QueryParams `json:"QueryParams"`
		PathParams  nats.PathParams  `json:"PathParams"`
	}{
		c.RequestId(),
		c.QueryParams(),
		c.PathParams(),
	}

	e, _ := json.Marshal(data)
	return nats.
		NewResBuilder().
		Body(e).
		Build()
}

func (h *Handler) Put(c *nats.Context, r *nats.Request) *nats.Response {
	return nats.
		NewResBuilder().
		Body(r.Body).
		Build()
}

type User struct {
	Name  string
	Email string
}

type Result struct {
	Name  string
	Email string
}

func (h *Handler) Post(c *nats.Context, user *User) (*Result, error) {
	return &Result{Name: user.Name + "_back", Email: user.Email + "_back"}, nil
}

func (h *Handler) Hello(c *nats.Context, r *nats.Request) *nats.Response {
	logger := c.Logger()
	logger.Info("Handler received request id " + c.RequestId())
	return nats.
		NewResBuilder().
		Body([]byte("hello world")).
		Build()
}
