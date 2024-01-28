package parser

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func evalStringExpression(expr hcl.Expression, success func(val string)) diagnostics.Diagnostics {
	val, d := expr.Value(nil)
	diags := convertHclDiagnostics(d)
	if diags.HasErrors() {
		return diags
	}
	if val.Type() != cty.String {
		r := expr.Range()
		diags = diags.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Unexpected type",
			Detail:   "Expected a string",
			Range:    convertHclRange(&r),
		})
		return diags
	}
	success(val.AsString())
	return diags
}
