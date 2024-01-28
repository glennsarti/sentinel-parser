package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseTestBlock(block *hcl.Block) (*ast.Test, diagnostics.Diagnostics) {
	content, d := block.Body.Content(TestBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := new(ast.Test)
	result.TestRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "rules":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.RulesRange = hclToSourceRange(&attr.Range)
			if !d.HasErrors() {
				for _, pair := range pairs {
					tr := ast.TestRule{
						TestRuleRange: hclToExplicitSourceRange(pair.Key.Range(), pair.Value.Range()),
					}
					// Key
					errs := evalStringExpression(pair.Key, func(key string) {
						tr.Name = key
						tr.NameRange = hclToSourceRange(pair.Key.Range().Ptr())
					})
					diags = append(diags, errs...)
					// Value
					v, e := pair.Value.Value(nil)
					if !e.HasErrors() {
						tr.ValueType = v.Type().FriendlyName()
						tr.ValueRange = hclToSourceRange(pair.Value.Range().Ptr())
					}
					diags = concatHclDiagnostics(diags, e)

					result.Rules = append(result.Rules, &tr)
				}
			}
			diags = concatHclDiagnostics(diags, d)
		}
	}

	return result, diags
}
