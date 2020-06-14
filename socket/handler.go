package socket

import (
	"net/http"

	"logur.dev/logur"
)

type HandlerFunc func(socketManager *SocketManager, logger logur.Logger, w http.ResponseWriter, r *http.Request)
