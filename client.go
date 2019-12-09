package titan

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	conn *Connection
}

func (srv *Client) request(ctx *Context, rq *Request, subject string) (*Response, error) {
	defer func(c *Connection) {
		_ = c.Conn.Flush()
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

	err = json.Unmarshal(msg.Body, &receive)
	if err != nil {
		return errors.WithMessage(err, "nats client json parsing error")
	}
	return nil
}

func (srv *Client) SendRequest(ctx *Context, rq *Request) (*Response, error) {
	t := time.Now()
	logger := ctx.Logger()
	// copy info inside context
	rq.Headers.Set(XRequestId, ctx.RequestId())
	//todo: copy authentication here

	subject := Url2Subject(rq.URL)
	logUrl := extractPartsFromUrl(rq.URL, 4, "/")

	logger.Debug("Nats client sending request", map[string]interface{}{
		"url": logUrl, "subject": subject, "id": ctx.RequestId(), "method": rq.Method})

	rp, err := srv.request(ctx, rq, subject)

	// just log event
	defer func() {
		var status string
		if err != nil {
			status = "error"
			logger.Error(fmt.Sprintf("Nats client receive error %+v", err), map[string]interface{}{
				"id":         ctx.RequestId(),
				"url":        logUrl,
				"status":     status,
				"subject":    subject,
				"elapsed_ms": float64(time.Since(t).Nanoseconds()) / 1000000.0},
			)
		} else {
			status = fmt.Sprintf("%d", rp.StatusCode)
		}
		logger.Debug("Nats client request complete", map[string]interface{}{
			"id":         ctx.RequestId(),
			"url":        logUrl,
			"status":     status,
			"subject":    subject,
			"elapsed_ms": float64(time.Since(t).Nanoseconds()) / 1000000.0},
		)
	}()

	if err != nil {
		var rpErr *Response
		if err.Error() == "nats: timeout" {
			rpErr = &Response{Status: "Request Timeout", StatusCode: 408}
		} else {
			rpErr = &Response{Status: "Internal Server Error", StatusCode: 500}
		}
		return nil, &ClientResponseError{Message: "Nats Client Request Timeout", Response: rpErr, Cause: err}
	}

	if rp.StatusCode >= 400 {
		return nil, &ClientResponseError{Message: rp.Status, Response: rp}
	}

	if rp.StatusCode >= 300 {
		return nil, &ClientResponseError{Message: "HTTP 3xx Redirection was not implemented yet", Response: rp}
	}

	if rp.StatusCode >= 200 {
		return rp, nil
	}

	if rp.StatusCode < 200 {
		return rp, &ClientResponseError{Message: "HTTP 1xx Informational response was not implemented yet", Response: rp}
	}

	return rp, nil
}

func extractPartsFromUrl(url string, numberOfPart int, separator string) string {
	if url == "" {
		return url
	}
	if !strings.Contains(url, "/") {
		return url
	}
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	s := strings.Split(url, "/")
	l := numberOfPart + 1
	if len(s) < l {
		l = len(s)
	}
	return strings.Join(s[1:l], separator)
}

func Url2Subject(url string) string {
	return extractPartsFromUrl(url, 3, ".")
}
