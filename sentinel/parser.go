package sentinel

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"

	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
	"github.com/glennsarti/sentinel-parser/sentinel/tracing"
)

type Parser interface {
	SentinelVersion() string

	ParseFile(
		lex Lexer,
		locator token.Locator,
		tracer tracing.ParsingTracer,
	) (*ast.File, diagnostics.Diagnostics, error)
}
