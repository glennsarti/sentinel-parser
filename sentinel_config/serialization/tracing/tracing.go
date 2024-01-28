package tracing

type Tracer interface {
	PushMethod(method, message string)
	Trace(message string)
	PopMethod(message string)
}

func NewNullTracer() Tracer {
	return nullTracer{}
}

type nullTracer struct{}

func (nt nullTracer) PushMethod(method, message string) {}
func (nt nullTracer) Trace(message string)              {}
func (nt nullTracer) PopMethod(message string)          {}
