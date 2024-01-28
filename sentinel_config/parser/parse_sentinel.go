package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"

	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseSentinelBlock(block *hcl.Block) (*ast.SentinelOptions, diagnostics.Diagnostics) {
	content, d := block.Body.Content(SentinelBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}

	result := &ast.SentinelOptions{
		Features: make([]*ast.Feature, 0),
	}
	result.SentinelOptionsRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "features":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.FeaturesRange = hclToSourceRange(&attr.Range)
			if !d.HasErrors() {
				for _, pair := range pairs {
					feat := ast.Feature{}
					// Key
					errs := evalStringExpression(pair.Key, func(key string) {
						feat.Name = key
						feat.NameRange = hclToSourceRange(pair.Key.Range().Ptr())
					})
					diags = append(diags, errs...)
					// Value
					v, e := pair.Value.Value(nil)
					if !e.HasErrors() {
						feat.Value = ast.ConvertCtyValue(&v)
						feat.ValueType = v.Type().FriendlyName()
						feat.ValueRange = hclToSourceRange(pair.Value.Range().Ptr())
					}
					diags = concatHclDiagnostics(diags, e)

					result.Features = append(result.Features, &feat)
				}
			}
			diags = concatHclDiagnostics(diags, d)
		}
	}

	return result, diags
}
