package controller

import "net/http"

func Hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello world"))
}
