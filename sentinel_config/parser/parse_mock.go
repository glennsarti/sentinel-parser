package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseMockBlock(block *hcl.Block) (*ast.Mock, diagnostics.Diagnostics) {
	content, d := block.Body.Content(MockBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.Mock{
		Data: make(map[string]*ast.Parameter, 0),
	}
	result.Name, result.NameRange = getLabelAndRange(block, 0, "")
	result.MockRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "data":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.DataRange = hclToSourceRange(&attr.Range)
			if !d.HasErrors() {
				for _, pair := range pairs {
					pa := ast.Parameter{}
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
					diags = diags.Concat(appendItem[ast.Parameter](result.Data, pa))
				}
			}
			diags = concatHclDiagnostics(diags, d)
		}
	}

	for _, bl := range content.Blocks {
		switch bl.Type {
		case "module":
			mod, e := parseMockModuleBlock(bl)
			if !e.HasErrors() {
				result.Module = mod
			}
			diags = diags.Concat(e)
		}
	}

	return result, diagnostics.EmptyDiags()
}

func parseMockModuleBlock(block *hcl.Block) (*ast.MockModule, diagnostics.Diagnostics) {
	content, d := block.Body.Content(MockModuleBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := new(ast.MockModule)
	result.MockModuleRange = hclToSourceRange(block.DefRange.Ptr())

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "source":
			d := evalStringExpression(attr.Expr, func(val string) {
				result.Source = val
				result.SourceRange = hclToSourceRange(&attr.Range)
			})
			diags = diags.Concat(d)
		}
	}

	return result, diags
}
