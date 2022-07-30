package tracing

var output Trace

func P(key string, val interface{}) Trace       { return output.P(key, val) }
func Debugf(format string, args ...interface{}) { output.Debugf(format, args...) }
func Infof(format string, args ...interface{})  { output.Infof(format, args...) }
func Errorf(format string, args ...interface{}) { output.Errorf(format, args...) }

type TB interface {
	Cleanup(func())
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}
