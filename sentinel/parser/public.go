package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/lexer"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
	"github.com/glennsarti/sentinel-parser/sentinel/tracing"
)

func New(sentinelVersion string) (sentinel.Parser, error) {
	if sentinelVersion == "" {
		sentinelVersion = features.LatestSentinelVersion
	}
	return newParser(sentinelVersion)
}

func ParseFileWithTracing(
	sentinelVersion,
	filename string,
	src []byte,
	tracer tracing.ParsingTracer,
) (*ast.File, token.Locator, diagnostics.Diagnostics, error) {
	loc := token.NewContentLocator(filename, src)

	lex, err := lexer.New(sentinelVersion, string(src))
	if err != nil {
		return nil, nil, nil, err
	}

	p, err := New(sentinelVersion)
	if err != nil {
		return nil, nil, nil, err
	}

	r, diags, err := p.ParseFile(lex, loc, tracer)

	return r, loc, diags, err
}

func ParseFile(
	sentinelVersion string,
	filename string,
	src []byte,
) (*ast.File, token.Locator, diagnostics.Diagnostics, error) {
	return ParseFileWithTracing(sentinelVersion, filename, src, nil)
}
