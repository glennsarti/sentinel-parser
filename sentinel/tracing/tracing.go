package tracing

type ParsingTracer interface {
	PushMethod(method, message string)
	Trace(message string)
	PopMethod(message string)
}

func NewNullTracer() ParsingTracer {
	return nullTracer{}
}

type nullTracer struct{}

func (nt nullTracer) PushMethod(method, message string) {}
func (nt nullTracer) Trace(message string)              {}
func (nt nullTracer) PopMethod(message string)          {}
