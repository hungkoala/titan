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

	"github.com/pkg/errors"
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

	logUrl := rq.URL

	logger.Debug("Nats client sending request to", map[string]interface{}{
		"url": logUrl, "target subject": subject, "id": requestId, "method": rq.Method})
	rp, err := srv.request(rq, subject)

	// just log event
	defer func(e error) {
		var status string
		if e != nil {
			status = "error"
			logger.Error(fmt.Sprintf("Nats client receive error %+v", err), map[string]interface{}{
				"id":             requestId,
				"err":            e.Error(),
				"url":            logUrl,
				"status":         status,
				"target subject": subject,
				"elapsed_ms":     float64(time.Since(t).Nanoseconds()) / 1000000.0},
			)
		} else {
			status = fmt.Sprintf("%d", rp.StatusCode)
		}
		logger.Debug("Nats client request complete", map[string]interface{}{
			"id":             requestId,
			"url":            logUrl,
			"status":         status,
			"target subject": subject,
			"elapsed_ms":     float64(time.Since(t).Nanoseconds()) / 1000000.0},
		)
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

	return rp, nil
}

func (srv *Client) Publish(ctx *Context, subject string, body interface{}) error {
	p := Message{
		Headers: http.Header{},
	}

	p.Headers.Set(XRequestId, ctx.RequestId())
	p.Headers.Set(XUserInfo, ctx.UserInfoJson())

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
