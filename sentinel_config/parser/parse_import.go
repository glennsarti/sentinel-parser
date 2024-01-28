package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	if len(block.Labels) == 1 {
		return parseV1PluginImportBlock(block)
	} else {
		return parseV2ImportBlock(block)
	}
}

func parseV1ModuleImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	content, d := block.Body.Content(v1ImportModuleBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.V1ModuleImport{}

	result.Name, result.NameRange = getLabelAndRange(block, 0, "")
	result.BlockRange = hclToSourceRange(&block.DefRange)

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

func parseV1PluginImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	content, d := block.Body.Content(v1ImportPluginBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.V1PluginImport{
		Config: make(map[string]*ast.Parameter, 0),
		Args:   make([]string, 0),
		Env:    make([]string, 0),
	}

	result.Name, result.NameRange = getLabelAndRange(block, 0, "")
	result.BlockRange = hclToSourceRange(&block.DefRange)

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "args":
			list, d := hcl.ExprList(attr.Expr)
			result.ArgsRange = hclToSourceRange(&attr.Range)
			if !d.HasErrors() {
				for _, item := range list {
					errs := evalStringExpression(item, func(val string) {
						result.Args = append(result.Args, val)
					})
					diags = diags.Concat(errs)
				}
			}
			diags = concatHclDiagnostics(diags, d)
		case "env":
			list, d := hcl.ExprList(attr.Expr)
			result.EnvRange = hclToSourceRange(&attr.Range)
			if !d.HasErrors() {
				for _, item := range list {
					errs := evalStringExpression(item, func(val string) {
						result.Env = append(result.Env, val)
					})
					diags = diags.Concat(errs)
				}
			}
			diags = concatHclDiagnostics(diags, d)
		case "config":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.ConfigRange = hclToSourceRange(&attr.Range)
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
					diags = diags.Concat(appendItem[ast.Parameter](result.Config, pa))
				}
			}
			diags = concatHclDiagnostics(diags, d)
		case "path":
			d := evalStringExpression(attr.Expr, func(val string) {
				result.Path = val
				result.PathRange = hclToSourceRange(&attr.Range)
			})
			diags = diags.Concat(d)
		}
	}

	return result, diags
}

func parseV2ImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	kind := getLabel(block, 0, "")
	switch kind {
	case ast.V2ImportKindModule:
		return parseV2ModuleImportBlock(block)
	case ast.V2ImportKindPlugin:
		return parseV2PluginImportBlock(block)
	case ast.V2ImportKindStatic:
		return parseV2StaticImportBlock(block)
	default:
		dr := block.DefRange
		return nil, diagnostics.Diagnostics{
			{
				Severity: diagnostics.Error,
				Summary:  fmt.Sprintf("Invalid import kind %s", kind),
				Detail: fmt.Sprintf("Invalid import kind %q, allowed values %s, %s or %s",
					kind,
					ast.V2ImportKindModule,
					ast.V2ImportKindPlugin,
					ast.V2ImportKindStatic,
				),
				Range: convertHclRange(&dr),
			},
		}
	}
}

func parseV2StaticImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	content, d := block.Body.Content(v2ImportStaticBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.V2StaticImport{}

	result.Kind, result.KindRange = getLabelAndRange(block, 0, "")
	result.Name, result.NameRange = getLabelAndRange(block, 1, "")
	result.BlockRange = hclToSourceRange(&block.DefRange)

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "format":
			d := evalStringExpression(attr.Expr, func(val string) {
				result.Format = val
				result.FormatRange = hclToSourceRange(&attr.Range)
			})
			diags = diags.Concat(d)
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

func parseV2PluginImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	content, d := block.Body.Content(v2ImportPluginBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.V2PluginImport{
		Config: make(map[string]*ast.Parameter, 0),
		Env:    make(map[string]*ast.Parameter, 0),
		Args:   make([]string, 0),
	}

	result.Kind, result.KindRange = getLabelAndRange(block, 0, "")
	result.Name, result.NameRange = getLabelAndRange(block, 1, "")
	result.BlockRange = hclToSourceRange(&block.DefRange)

	for _, attr := range content.Attributes {
		switch attr.Name {
		case "args":
			list, d := hcl.ExprList(attr.Expr)
			result.ArgsRange = hclToSourceRange(&attr.Range)
			if !d.HasErrors() {
				for _, item := range list {
					errs := evalStringExpression(item, func(val string) {
						result.Args = append(result.Args, val)
					})
					diags = diags.Concat(errs)
				}
			}
			diags = concatHclDiagnostics(diags, d)
		case "env":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.EnvRange = hclToSourceRange(&attr.Range)
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
					diags = append(diags, appendItem[ast.Parameter](result.Env, pa)...)
				}
			}
			diags = concatHclDiagnostics(diags, d)
		case "config":
			pairs, d := hcl.ExprMap(attr.Expr)
			result.ConfigRange = hclToSourceRange(&attr.Range)
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
					diags = diags.Concat(appendItem[ast.Parameter](result.Config, pa))
				}
			}
			diags = concatHclDiagnostics(diags, d)
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

func parseV2ModuleImportBlock(block *hcl.Block) (ast.Import, diagnostics.Diagnostics) {
	content, d := block.Body.Content(v2ImportModuleBlockSchema)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return nil, diags
	}
	result := &ast.V2ModuleImport{}

	result.Kind, result.KindRange = getLabelAndRange(block, 0, "")
	result.Name, result.NameRange = getLabelAndRange(block, 1, "")
	result.BlockRange = hclToSourceRange(&block.DefRange)

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
