package tracing

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics"

	jaegercfg "github.com/uber/jaeger-client-go/config"
)

const (
	XRequestId          = "X-Request-Id"
	UberTraceID         = "Uber-Trace-Id"
	SERVICE_NAME        = "SERVICE_NAME"
	JAEGER_SERVICE_NAME = "JAEGER_SERVICE_NAME"
)

var one sync.Once
var tracer opentracing.Tracer

func init() {
	viper.SetDefault(
		"SERVICE_NAME",
		fmt.Sprintf("Default_titan_client_%s", uuid.New().String()))
}

func InitTracing(serviceName string) {
	tracer = initTracing(serviceName)
}

func initTracing(serviceName string) opentracing.Tracer {
	os.Setenv(JAEGER_SERVICE_NAME, serviceName)

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
		return nil
	}

	jLogger := newLogger()
	jMetricsFactory := metrics.NullFactory

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("couldn't setup tracing: %v \n", err)
		return nil
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(close io.Closer) {
		<-c
		log.Println("\r- Ctrl+C pressed in Terminal")
		close.Close()
	}(closer)

	return tracer
}

func GetTracer() opentracing.Tracer {
	if tracer == nil {
		one.Do(func() {
			tracer = initTracing(viper.GetString(SERVICE_NAME))
		})
	}
	return tracer
}

func SpanContext(
	carier *http.Header,
	url string,
) opentracing.Span {
	if tracer == nil {
		return nil
	}
	m := make(opentracing.TextMapCarrier)
	m.Set(UberTraceID, carier.Get(UberTraceID))
	spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, m)
	var reqSpan opentracing.Span

	urls := strings.Split(url, ".")
	if len(urls) > 0 {
		url = urls[len(urls)-1]
	}
	if err == nil {
		reqSpan = tracer.StartSpan(
			url,
			opentracing.ChildOf(spanCtx),
			ext.SpanKindRPCServer,
			ext.RPCServerOption(spanCtx))
	} else {
		reqSpan = tracer.StartSpan(url, ext.SpanKindRPCClient)
	}

	reqSpan.SetTag(XRequestId, carier.Get(XRequestId))
	if err := tracer.Inject(reqSpan.Context(), opentracing.HTTPHeaders, carier); err != nil {
		log.Printf("%v for Inject.", err)
	}
	return reqSpan
}
