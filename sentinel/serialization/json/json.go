package json

import (
	"io"

	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/serialization"
)

type jsonCoder struct {
	tracer Tracer
}

func New(tracer Tracer) serialization.Serializer {
	if tracer == nil {
		tracer = NewNullTracer()
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
