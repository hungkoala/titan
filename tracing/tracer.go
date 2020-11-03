package tracing

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics"

	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
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
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
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
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		close.Close()
	}(closer)

	return tracer
}
func GetTracer() opentracing.Tracer {
	if tracer == nil {
		one.Do(func() {
			name := viper.GetString("SERVICE_NAME")
			tracer = initTracing(name)
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
	m := make(opentracing.TextMapCarrier, 0)
	m.Set("Uber-Trace-Id", carier.Get("Uber-Trace-Id"))
	spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, m)
	var reqSpan opentracing.Span
	if err == nil {
		reqSpan = tracer.StartSpan(
			url,
			opentracing.ChildOf(spanCtx),
			ext.SpanKindRPCServer,
			ext.RPCServerOption(spanCtx))
	} else {
		reqSpan = tracer.StartSpan(url, ext.SpanKindRPCClient)
	}

	ext.MessageBusDestination.Set(reqSpan, url)
	if err := tracer.Inject(reqSpan.Context(), opentracing.HTTPHeaders, carier); err != nil {
		log.Printf("%v for Inject.", err)
	}
	return reqSpan
}
