package sentinel_config

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"

	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

type Parser interface {
	SentinelVersion() string

	ParseFile(filename string, src []byte) (*ast.File, diagnostics.Diagnostics)
}
