package tracing

import (
	"log"
	"os"
)

type jaegerLogger struct {
	loggerOut *log.Logger
}

func newLogger() *jaegerLogger {
	return &jaegerLogger{
		loggerOut: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (t *jaegerLogger) Error(msg string) {
	t.loggerOut.Println("Err: ", msg)
}

func (t *jaegerLogger) Infof(msg string, args ...interface{}) {
	//t.loggerOut.Println(msg, args)
}
