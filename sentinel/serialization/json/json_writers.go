package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
)

func (jc *jsonCoder) writeHeader(typeName string, writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `{"_t":"%s","_p":{`, typeName)
	return err
}

func (jc *jsonCoder) writeFooter(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `}}`)
	return err
}

func (jc *jsonCoder) writeLiteralString(value string, writer io.Writer) error {
	if v, err := json.Marshal(value); err != nil {
		return err
	} else {
		_, err := writer.Write(v)
		return err
	}
}

func (jc *jsonCoder) writeLiteralBool(value bool, writer io.Writer) error {
	if v, err := json.Marshal(value); err != nil {
		return err
	} else {
		_, err := writer.Write(v)
		return err
	}
}

func (jc *jsonCoder) writeSourceRange(src position.SourceRange, writer io.Writer) error {
	if v, err := json.Marshal(jsonSourceRange{
		End: jsonSourcePos{
			Byte:   src.End.Byte,
			Column: src.End.Column,
			Line:   src.End.Line,
		},
		Filename: src.Filename,
		Start: jsonSourcePos{
			Byte:   src.Start.Byte,
			Column: src.Start.Column,
			Line:   src.Start.Line,
		},
	}); err != nil {
		return err
	} else {
		_, err := writer.Write(v)
		return err
	}
}

func (jc *jsonCoder) writeSourceRangePointer(src *position.SourceRange, writer io.Writer) error {
	if src == nil {
		fmt.Fprintf(writer, "null")
		return nil
	} else {
		return jc.writeSourceRange(*src, writer)
	}
}

func (jc *jsonCoder) writeNilableNode(src ast.Node, writer io.Writer) error {
	if src == nil {
		fmt.Fprintf(writer, "null")
		return nil
	} else {
		return jc.write(src, writer)
	}
}

func writeList[T ast.Node](obj []T, jc *jsonCoder, writer io.Writer) error {
	if _, err := fmt.Fprint(writer, `[`); err != nil {
		return err
	}
	for idx, item := range obj {
		if idx > 0 {
			if _, err := fmt.Fprint(writer, `,`); err != nil {
				return err
			}
		}
		if err := jc.write(item, writer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(writer, `]`)
	return err
}
