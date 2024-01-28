package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseParamBlock(block *hcl.Block) (*ast.Parameter, diagnostics.Diagnostics) {
	content, diags := block.Body.Content(ParamBlockSchema)
	if diags.HasErrors() {
		return nil, convertHclDiagnostics(diags)
	}
	result := new(ast.Parameter)
	result.Name, result.NameRange = getLabelAndRange(block, 0, "")
	result.ParameterRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "value":
			v, errs := attr.Expr.Value(nil)
			if !errs.HasErrors() {
				result.Value = ast.ConvertCtyValue(&v)
				result.ValueType = v.Type().FriendlyName()
				result.ValueRange = hclToSourceRange(attr.Range.Ptr())
			}
			diags = append(diags, errs...)
		}
	}

	return result, diagnostics.EmptyDiags()
}
