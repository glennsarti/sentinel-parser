package parser

import (
	"fmt"
	"strings"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type parser struct {
	p       *hclparse.Parser
	version string
}

func newParser(sentinelVersion string) (*parser, error) {
	ok, ver := features.ValidateSentinelVersion(sentinelVersion)
	if !ok {
		return nil, fmt.Errorf("%s is an invalid sentinel version", sentinelVersion)
	}

	return &parser{
		p:       hclparse.NewParser(),
		version: ver,
	}, nil
}

func (p *parser) SentinelVersion() string {
	return p.version
}

func (p *parser) ParseFile(filename string, src []byte) (*ast.File, diagnostics.Diagnostics) {
	// Parse the HCL content
	file, d := p.parseFileSource(src, filename)
	diags := convertHclDiagnostics(d)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}

	if file == nil || file.Body == nil {
		diags = diags.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Failed to parse file",
			Detail:   fmt.Sprintf("The file %q could not be parsed.", filename),
		})
		return nil, diags
	}

	// Parse the HCL content into a Sentinel Configuration File
	configFile, e := parseConfigFile(file.Body, p.version)
	diags = diags.Concat(e)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}
	if configFile == nil {
		diags = diags.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Failed to parse file",
			Detail:   fmt.Sprintf("The file %q could not be parse.", filename),
		})
		return nil, diags
	}

	return configFile, diags
}

func (p *parser) parseFileSource(src []byte, path string) (*hcl.File, hcl.Diagnostics) {
	switch {
	case strings.HasSuffix(path, ".json"):
		return p.p.ParseJSON(src, path)
	default:
		return p.p.ParseHCL(src, path)
	}
}
