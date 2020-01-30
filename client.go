package titan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	conn *Connection
}

var null = []byte{'n', 'u', 'l', 'l'}

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

	if bytes.Equal(msg.Body, null) {
		receive = nil
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
		return s[0:len(s) - 1]
	}
	return s
}

func (srv *Client) SendRequest(ctx *Context, rq *Request) (*Response, error) {
	t := time.Now()
	logger := ctx.Logger()
	// copy info inside context
	rq.Headers.Set(XRequestId, ctx.RequestId())

	//todo: copy authentication here
	userInfoJson := ctx.UserInfoJson()
	if userInfoJson != "" {
		rq.Headers.Set(XUserInfo, ctx.UserInfoJson())
	}

	// hacked code,  dont know why we need it, should re-check it in micronaut
	multiTenantCareProviderId, ok := ctx.Value("multiTenantCareProviderId").(string)
	if !ok && multiTenantCareProviderId != "" {
		rq.Headers.Set("multiTenantCareProviderId", multiTenantCareProviderId)
	}
	// end of hacked code

	subject := Url2Subject(rq.URL)
	logUrl := extractLoggablePartsFromUrl(rq.URL)

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

	if rp.StatusCode < 200 {
		return nil, &ClientResponseError{Message: "HTTP 1xx Informational response was not implemented yet", Response: rp}
	}

	return rp, nil
}
