package titan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/silenteer-oss/titan/log"

	"logur.dev/logur"
)

type Middleware struct {
	//name   string
	//logger logur.Logger
}

func NewMiddleware(name string, subject string, logger logur.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			t := time.Now()
			if r.Header == nil {
				r.Header = http.Header{}
			}
			requestID := r.Header.Get(XRequestId)
			if requestID == "" {
				requestID = RandomString(6)
				r.Header.Set(XRequestId, requestID)
			}
			logWithId := log.WithFields(logger, map[string]interface{}{
				"id":      requestID,
				"method":  r.Method,
				"subject": subject,
				"url":     r.URL.Path,
			})

			url := ExtractLoggablePartsFromUrl(r.URL.Path)
			logWithId.Debug(name+" server received request", map[string]interface{}{"url": url})

			ctx := r.Context()

			// add log
			ctx = context.WithValue(ctx, XLoggerId, logWithId)

			// add request id
			ctx = context.WithValue(ctx, XRequestId, requestID)

			origin := r.Header.Get(XOrigin)
			ctx = context.WithValue(ctx, XOrigin, origin)

			//add user info
			userInfoJson := r.Header.Get(XUserInfo)
			if userInfoJson != "" {
				var userInfo UserInfo
				err := json.Unmarshal([]byte(userInfoJson), &userInfo)
				if err != nil {
					logger.Error(fmt.Sprintf("Unmarshal User Info  error: %+v\n ", err))
				} else {
					ctx = context.WithValue(ctx, XUserInfo, &userInfo)
				}
			}

			// add query params
			queryParams := QueryParams(r.URL.Query())
			ctx = context.WithValue(ctx, XQueryParams, queryParams)

			rp := NewCustomResponseWriter(w)

			defer func() {
				logWithId.Debug(name+" server request complete", map[string]interface{}{
					"status":     rp.StatusCode,
					"elapsed_ms": float64(time.Since(t).Nanoseconds()) / 1000000.0},
				)
			}()

			next.ServeHTTP(rp, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

type CustomResponseWriter struct {
	w          http.ResponseWriter
	StatusCode int
}

func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{w: w}
}

func (c *CustomResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CustomResponseWriter) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *CustomResponseWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
	c.StatusCode = statusCode
}
