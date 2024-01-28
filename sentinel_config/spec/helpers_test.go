package spec

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

func diffCompareOptions() cmp.Options {
	return cmp.Options{
		cmp.Comparer(ctyValueComparer),
		cmp.Comparer(ctyValuePtrComparer),
		cmp.Comparer(dynamicValueComparer),
		cmp.Comparer(dynamicValuePtrComparer),
	}
}

func ctyValueComparer(x, y cty.Value) bool {
	return x.GoString() == y.GoString()
}

func ctyValuePtrComparer(x, y *cty.Value) bool {
	if x == nil && y == nil {
		return true
	}
	if x != nil && y != nil {
		return x.GoString() == y.GoString()
	}
	return false
}

func dynamicValueComparer(x, y ast.DynamicValue) bool {
	return x.GoString() == y.GoString()
}

func dynamicValuePtrComparer(x, y *ast.DynamicValue) bool {
	if x == nil && y == nil {
		return true
	}
	if x != nil && y != nil {
		return x.GoString() == y.GoString()
	}
	return false
}

func rangeToString(r *position.SourceRange) string {
	if r == nil {
		return ":nil"
	}
	return fmt.Sprintf(":%s:%d,%d->%d,%d", r.Filename, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
}
