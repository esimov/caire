//go:build uaxtrace
// +build uaxtrace

package tracing

import (
	"fmt"
	"log"
	"strings"
)

type Trace struct {
	prefix string
	log    *log.Logger
}

func init() { output.log = log.Default() }

func SetTestingLog(tb TB) {
	output.log = log.New(tbWriter{tb}, "", log.LstdFlags)
	tb.Cleanup(func() { output.log = log.Default() })
}

func (trace Trace) P(key string, val interface{}) Trace {
	return Trace{
		prefix: trace.prefix + fmt.Sprintf("[%s=%v] ", key, val),
		log:    trace.log,
	}
}

func (trace Trace) Debugf(format string, args ...interface{}) {
	trace.output("DEBUG ", "", format, args...)
}

func (trace Trace) Infof(format string, args ...interface{}) {
	trace.output("INFO ", "", format, args...)
}

func (trace Trace) Errorf(format string, args ...interface{}) {
	trace.output("ERROR ", "", format, args...)
}

func (trace Trace) output(prefix string, p string, s string, args ...interface{}) {
	trace.log.SetPrefix(prefix)
	if p == "" { // if no prefix present
		trace.log.Printf(s, args...)
	} else {
		trace.log.Println(p + fmt.Sprintf(s, args...))
	}
}

type tbWriter struct{ tb TB }

func (w tbWriter) Write(data []byte) (int, error) {
	w.tb.Log(strings.TrimSuffix(string(data), "\n"))
	return len(data), nil
}
