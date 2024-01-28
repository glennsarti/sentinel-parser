package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

func asPointer[E any](item E) *E {
	return &item
}

func appendItem[T ast.HCLNode](list map[string]*T, item T) diagnostics.Diagnostics {
	if src, ok := list[item.BlockName()]; ok {
		return diagnostics.Diagnostics{
			{
				Severity: diagnostics.Error,
				Summary:  fmt.Sprintf("Duplicate %s block defined", item.BlockType()),
				Detail: fmt.Sprintf("%s block with the name of %q has already been defined.",
					item.BlockType(),
					item.BlockName(),
				),
				Range: ast.CloneRange(item.BlockNameRange()),
				Related: &[]diagnostics.RelatedDiagnostic{
					{
						Summary: fmt.Sprintf("First %s block definition with name %q.", item.BlockType(), item.BlockName()),
						Range:   ast.CloneRange((*src).BlockNameRange()),
					},
				},
			},
		}
	}
	list[item.BlockName()] = asPointer(item)
	return diagnostics.EmptyDiags()
}

func appendImport(m map[string]ast.Import, item ast.Import) diagnostics.Diagnostics {
	if src, ok := m[item.BlockName()]; ok {
		blockType := item.BlockType()
		if item.Schema() == ast.V2ImportSchema {
			blockType = "import"
		}
		return diagnostics.Diagnostics{
			{
				Severity: diagnostics.Error,
				Summary:  fmt.Sprintf("Duplicate %s block defined", blockType),
				Detail:   fmt.Sprintf("%s block with the name of %q has already been defined.", blockType, item.BlockName()),
				Range:    ast.CloneRange(item.BlockNameRange()),
				Related: &[]diagnostics.RelatedDiagnostic{
					{
						Summary: fmt.Sprintf("First block definition with name %q.", item.BlockName()),
						Range:   ast.CloneRange(src.BlockNameRange()),
					},
				},
			},
		}
	}
	m[item.BlockName()] = item
	return diagnostics.EmptyDiags()
}
