//go:build !uaxtrace
// +build !uaxtrace

package tracing

type Trace struct{}

func SetTestingLog(tb TB) {}

func (trace Trace) P(key string, val interface{}) Trace       { return Trace{} }
func (trace Trace) Debugf(format string, args ...interface{}) {}
func (trace Trace) Infof(format string, args ...interface{})  {}
func (trace Trace) Errorf(format string, args ...interface{}) {}
