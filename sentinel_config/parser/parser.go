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

type Parser struct {
	p       *hclparse.Parser
	version string
}

func NewParser(sentinelVersion string) (*Parser, error) {
	ok, ver := features.ValidateSentinelVersion(sentinelVersion)
	if !ok {
		return nil, fmt.Errorf("%s is an invalid sentinel version", sentinelVersion)
	}

	return &Parser{
		p:       hclparse.NewParser(),
		version: ver,
	}, nil
}

func (p *Parser) ParseFileSource(src []byte, path string) (*ast.File, diagnostics.Diagnostics) {
	// Parse the HCL content
	file, d := p.parseFileSource(src, path)
	diags := convertHclDiagnostics(d)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}

	if file == nil || file.Body == nil {
		diags = diags.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Failed to parse file",
			Detail:   fmt.Sprintf("The file %q could not be parsed.", path),
		})
		return nil, diags
	}

	// Parse the HCL content into a Sentinel File
	configFile, e := parseConfigFile(file.Body, p.version)
	diags = diags.Concat(e)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}
	if configFile == nil {
		diags = diags.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Failed to parse file",
			Detail:   fmt.Sprintf("The file %q could not be parse.", path),
		})
		return nil, diags
	}

	return configFile, diags
}

func (p *Parser) parseFileSource(src []byte, path string) (*hcl.File, hcl.Diagnostics) {
	switch {
	case strings.HasSuffix(path, ".json"):
		return p.p.ParseJSON(src, path)
	default:
		return p.p.ParseHCL(src, path)
	}
}
