package serialization

import (
	"io"

	"github.com/glennsarti/sentinel-parser/sentinel/ast"
)

type Serializer interface {
	Serialize(ast.Node, io.Writer) error
	Deserialize(io.Reader, ast.Node) error
}
