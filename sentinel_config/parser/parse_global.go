package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseGlobalBlock(block *hcl.Block) (*ast.Global, diagnostics.Diagnostics) {
	content, diags := block.Body.Content(GlobalBlockSchema)
	if diags.HasErrors() {
		return nil, convertHclDiagnostics(diags)
	}
	result := new(ast.Global)
	result.Name, result.NameRange = getLabelAndRange(block, 0, "")
	result.GlobalRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "value":
			v, errs := attr.Expr.Value(nil)
			if !errs.HasErrors() {
				result.Value = ast.ConvertCtyValue(&v)
				result.ValueType = v.Type().FriendlyName()
				result.ValueRange = hclToSourceRange(attr.Expr.Range().Ptr())
			}
			diags = append(diags, errs...)
		}
	}

	return result, diagnostics.EmptyDiags()
}
