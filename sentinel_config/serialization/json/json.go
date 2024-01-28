package json

import (
	"io"

	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/glennsarti/sentinel-parser/sentinel_config/serialization"
	"github.com/glennsarti/sentinel-parser/sentinel_config/serialization/tracing"
)

type jsonCoder struct {
	tracer tracing.Tracer
}

func New(tracer tracing.Tracer) serialization.Serializer {
	if tracer == nil {
		tracer = tracing.NewNullTracer()
	}
	return &jsonCoder{
		tracer: tracer,
	}
}

func (jc *jsonCoder) Serialize(node ast.Node, writer io.Writer) error {
	if err := jc.write(node, writer); err != nil {
		return err
	}

	return nil
}

func (jc *jsonCoder) Deserialize(r io.Reader, into ast.Node) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return jc.readInto(data, into)
}
