package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel_config"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

var _ sentinel_config.Parser = &parser{}

func New(sentinelVersion string) (sentinel_config.Parser, error) {
	if sentinelVersion == "" {
		sentinelVersion = features.LatestSentinelVersion
	}
	return newParser(sentinelVersion)
}

func ParseFile(
	sentinelVersion,
	filename string,
	src []byte,
) (*ast.File, diagnostics.Diagnostics, error) {
	p, err := New(sentinelVersion)
	if err != nil {
		return nil, nil, err
	}

	r, diags := p.ParseFile(filename, src)

	return r, diags, nil
}
