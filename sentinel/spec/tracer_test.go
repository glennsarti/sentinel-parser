package spec

import (
	"strings"
	"testing"

	"github.com/glennsarti/sentinel-parser/sentinel/tracing"
)

func NewSpecTracer(t *testing.T) tracing.ParsingTracer {
	return &specTracer{t: t}
}

type specTracer struct {
	t      *testing.T
	indent int
}

func (st *specTracer) PushMethod(method, message string) {
	st.t.Logf("%s(%s) BEGIN %s",
		strings.Repeat("| ", st.indent),
		method,
		message,
	)
	st.indent += 1
}
func (st *specTracer) Trace(message string) {
	st.t.Logf("%s%s",
		strings.Repeat("| ", st.indent),
		message,
	)
}
func (st *specTracer) PopMethod(message string) {
	st.indent -= 1
	st.t.Logf("%s(%s) END %s",
		strings.Repeat("| ", st.indent),
		"???",
		message,
	)
}
