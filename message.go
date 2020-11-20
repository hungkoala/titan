package titan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type Message struct {
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

func (r *Message) bodyJson(v interface{}) error {
	if r.Body == nil {
		return errors.New("body not found")
	}
	if err := json.Unmarshal(r.Body, &v); err != nil {
		return errors.WithMessage(err, "Json Unmarshal error ")
	}
	return nil
}

func (r *Message) context() (*Context, error) {
	ctx := context.Background()

	userInfoJson := r.Headers.Get(XUserInfo)

	if userInfoJson != "" {
		var userInfo UserInfo
		jerr := json.Unmarshal([]byte(userInfoJson), &userInfo)
		if jerr != nil {
			logger.Error(fmt.Sprintf("Unmarshal User Info  error: %+v\n ", jerr))
			return nil, jerr
		} else {
			ctx = context.WithValue(ctx, XUserInfo, &userInfo)
		}
	}

	ctx = context.WithValue(ctx, XRequestId, r.Headers.Get(XRequestId))
	ctx = context.WithValue(ctx, XOrigin, r.Headers.Get(XOrigin))
	ctx = context.WithValue(ctx, UberTraceID, r.Headers.Get(UberTraceID))

	return NewContext(ctx), nil
}

func (r *Message) Parse(v interface{}) (*Context, error) {
	err := r.bodyJson(v)
	if err != nil {
		return nil, err
	}

	return r.context()
}
