package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parsePolicyBlock(block *hcl.Block) (*ast.Policy, diagnostics.Diagnostics) {
	content, d := block.Body.Content(PolicyBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.Policy{
		Params: make(map[string]*ast.Parameter, 0),
	}
	result.Name, result.NameRange = getLabelAndRange(block, 0, "")
	result.PolicyRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "source":
			d := evalStringExpression(attr.Expr, func(val string) {
				result.Source = val
				result.SourceRange = hclToSourceRange(&attr.Range)
			})
			diags = diags.Concat(d)

		case "enforcement_level":
			d := evalStringExpression(attr.Expr, func(val string) {
				result.EnforcementLevel = val
				result.EnforcementLevelRange = hclToSourceRange(&attr.Range)
			})
			diags = diags.Concat(d)

		case "params":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.ParamsRange = hclToSourceRange(&attr.Range)
			result.Params = make(map[string]*ast.Parameter, 0)
			if !d.HasErrors() {
				for _, pair := range pairs {
					pa := ast.Parameter{
						ParameterRange: hclToExplicitSourceRange(pair.Key.Range(), pair.Value.Range()),
					}
					// Key
					errs := evalStringExpression(pair.Key, func(key string) {
						pa.Name = key
						pa.NameRange = hclToSourceRange(pair.Key.Range().Ptr())
					})
					diags = diags.Concat(errs)
					// Value
					v, e := pair.Value.Value(nil)
					if !e.HasErrors() {
						pa.Value = ast.ConvertCtyValue(&v)
						pa.ValueType = v.Type().FriendlyName()
						pa.ValueRange = hclToSourceRange(pair.Value.Range().Ptr())
					}
					diags = concatHclDiagnostics(diags, e)

					diags = diags.Concat(appendItem[ast.Parameter](result.Params, pa))
				}
			}
			diags = convertHclDiagnostics(d)
		}
	}

	return result, diagnostics.EmptyDiags()
}
