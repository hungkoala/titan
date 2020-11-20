package titan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.com/silenteer-oss/titan/tracing"
)

type Client struct {
	conn IConnection
}

func NewClient(conn IConnection) *Client {
	return &Client{conn: conn}
}

var null = []byte{'n', 'u', 'l', 'l'}

func (srv *Client) request(rq *Request, subject string) (*Response, error) {
	defer func(c IConnection) {
		_ = c.Flush()
	}(srv.conn)

	return srv.conn.SendRequest(rq, subject)
}

func (srv *Client) SendAndReceiveJson(ctx *Context, rq *Request, receive interface{}) error {
	msg, err := srv.SendRequest(ctx, rq)
	if err != nil {
		return err
	}

	if msg.Body == nil || len(msg.Body) == 0 {
		return nil
	}

	if bytes.Equal(msg.Body, null) {
		//implement receive = nil
		v := reflect.ValueOf(receive)
		v.Elem().Set(reflect.Zero(v.Elem().Type()))
		return nil
	}
	// very very stupid code, keep it here because micronaut used it.  please return json instead
	switch v := receive.(type) {
	case *string:
		ctx.Logger().Trace("expected type ", map[string]interface{}{"type": v})
		ptr := receive.(*string)
		*ptr = string(msg.Body)
	case *int:
		result, err := strconv.ParseInt(cleanString(msg.Body), 10, 0)
		if err != nil {
			return errors.WithMessage(err, "paring integer error")
		}
		ptr := receive.(*int)
		*ptr = int(result)
	case *int32:
		result, err := strconv.ParseInt(cleanString(msg.Body), 10, 32)
		if err != nil {
			return errors.WithMessage(err, "paring int32 error")
		}
		ptr := receive.(*int32)
		*ptr = int32(result)
	case *int64:
		result, err := strconv.ParseInt(cleanString(msg.Body), 10, 64)
		if err != nil {
			return errors.WithMessage(err, "paring int64 error")
		}
		ptr := receive.(*int64)
		*ptr = result
	case *uint:
		result, err := strconv.ParseUint(cleanString(msg.Body), 10, 0)
		if err != nil {
			return errors.WithMessage(err, "paring uint error")
		}
		ptr := receive.(*uint)
		*ptr = uint(result)
	case *float64:
		result, err := strconv.ParseFloat(cleanString(msg.Body), 64)
		if err != nil {
			return errors.WithMessage(err, "paring float64 error")
		}
		ptr := receive.(*float64)
		*ptr = result
	case *float32:
		result, err := strconv.ParseFloat(cleanString(msg.Body), 32)
		if err != nil {
			return errors.WithMessage(err, "paring float32 error")
		}
		ptr := receive.(*float32)
		*ptr = float32(result)
	case *bool:
		result, err := strconv.ParseBool(cleanString(msg.Body))
		if err != nil {
			return errors.WithMessage(err, "paring bool error")
		}
		ptr := receive.(*bool)
		*ptr = result
	default:
		err := json.Unmarshal(msg.Body, &receive)
		if err != nil {
			return errors.WithMessage(err, "nats client json parsing error")
		}
	}
	return nil
}

// cleanString clean \n in the last of string to avoid error when using strconv.Parse
// when lib marshal byte to json, it adds '\n\ in the value which causes error for strconv.Parse.
//
func cleanString(str []byte) string {
	s := string(str)
	if strings.HasSuffix(s, "\n") {
		return s[0 : len(s)-1]
	}
	return s
}

func (srv *Client) SendRequest(ctx *Context, rq *Request) (*Response, error) {
	t := time.Now()
	logger := ctx.Logger()

	// build request id
	requestId := ctx.RequestId()
	if requestId == "" && rq.Headers != nil {
		requestId = rq.Headers.Get(XRequestId)
	}


	if requestId == "" {
		requestId = RandomString(6)
	}
	rq.Headers.Set(XRequestId, requestId)

	origin := ctx.Origin()
	if origin == "" && rq.Headers != nil {
		origin = rq.Headers.Get(XOrigin)
	}

	rq.Headers.Set(XOrigin, origin)

	rq.Headers.Set(XRequestTime, strconv.FormatInt(time.Now().UnixNano(), 10))

	uberTraceID := ctx.UberTraceID()
	rq.Headers.Set(UberTraceID, uberTraceID)
	reqSpan := tracing.SpanContext(&rq.Headers, rq.URL)
	if reqSpan != nil {
		defer reqSpan.Finish()
	}

	//todo: copy authentication here
	userInfoJson := ctx.UserInfoJson()
	if userInfoJson != "" {
		rq.Headers.Set(XUserInfo, ctx.UserInfoJson())
	}

	// end of hacked code
	subject := rq.Subject
	if rq.Subject == "" {
		subject = Url2Subject(rq.URL)
	}

	logger.Debug("Nats client sending request to", map[string]interface{}{"url": rq.URL, "id": requestId, "method": rq.Method})
	rp, err := srv.request(rq, subject)

	// just log event
	defer func(e error) {
		elapsedMs := float64(time.Since(t).Nanoseconds()) / 1000000.0
		logInfo := map[string]interface{}{
			"id":         requestId,
			"url":        rq.URL,
			"elapsed_ms": elapsedMs,
			"status":     fmt.Sprintf("%d", rp.StatusCode),
		}

		if e != nil {
			logInfo["status"] = "error"
			logInfo["err"] = e.Error()

			if reqSpan != nil {
				ext.LogError(reqSpan, err)
			}
		}

		if float64(time.Since(t).Nanoseconds())/1000000.0 > 2000 {
			logger.Warn("Nats client slow api detected ", logInfo)
		} else {
			logger.Debug("Nats client request complete", logInfo)
		}
	}(err)

	// server return error object
	if err != nil {
		var rpErr *Response
		headers := http.Header{}
		headers.Add(XRequestId, requestId)

		if err.Error() == "nats: timeout" {
			rpErr = &Response{Status: "Request Timeout :" + requestId, StatusCode: 408, Headers: headers}
		} else {
			rpErr = &Response{Status: "Internal Server Error: " + requestId, StatusCode: 500, Headers: headers}
		}
		return nil, &ClientResponseError{Message: err.Error(), Response: rpErr, Cause: err}
	}

	// server return status code
	if rp.Headers == nil {
		rp.Headers = http.Header{}
	}
	if rp.Headers.Get(XRequestId) == "" {
		rp.Headers.Add(XRequestId, requestId)
	}

	if rp.StatusCode >= 400 {
		return nil, &ClientResponseError{Message: rp.Status, Response: rp}
	}

	if rp.StatusCode >= 300 {
		return nil, &ClientResponseError{Message: "HTTP 3xx Redirection was not implemented yet", Response: rp}
	}

	if rp.StatusCode < 200 {
		return nil, &ClientResponseError{Message: "HTTP 1xx Informational response was not implemented yet", Response: rp}
	}

	// log  too high latency
	requestTimeStr := rp.Headers.Get(XResponeTime)
	if requestTimeStr != "" {
		requestTime, err := strconv.ParseInt(requestTimeStr, 10, 64)
		if err == nil {
			duration := (time.Now().UnixNano() - requestTime) / 1000000
			if duration > 2000 { // 1 second
				logger.Warn("response latency is too high", map[string]interface{}{"time": duration})
			}
		}
	}

	rp.Headers.Set(XResponeTime, "")

	return rp, nil
}

func (srv *Client) Publish(ctx *Context, subject string, body interface{}) error {
	p := Message{
		Headers: http.Header{},
	}

	p.Headers.Set(XRequestId, ctx.RequestId())
	p.Headers.Set(XOrigin, ctx.Origin())
	p.Headers.Set(XUserInfo, ctx.UserInfoJson())

	p.Headers.Set(UberTraceID, ctx.UberTraceID())
	uberTraceID := ctx.UberTraceID()
	p.Headers.Set(UberTraceID, uberTraceID)
	reqSpan := tracing.SpanContext(&p.Headers, subject)
	if reqSpan != nil {
		defer reqSpan.Finish()
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	p.Body = b

	return srv.conn.Publish(subject, p)
}

func (srv *Client) Subscribe(subject string, cb Handler) (ISubscription, error) {
	return srv.conn.Subscribe(subject, cb)
}
