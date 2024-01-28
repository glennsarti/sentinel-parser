package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/hashicorp/hcl/v2"
)

func parseConfigFile(body hcl.Body, sentinelVersion string) (*ast.File, diagnostics.Diagnostics) {
	file := ast.NewFile()
	schema, err := buildHclSchema(body, sentinelVersion)
	if err != nil {
		return nil, errorToDiagnostics(err)
	}

	content, d := body.Content(schema)
	diags := convertHclDiagnostics(d)
	for _, block := range content.Blocks {
		switch block.Type {
		case "global":
			result, e := parseGlobalBlock(block)
			if result != nil {
				e = e.Concat(appendItem(file.Globals, *result))
			}
			diags = diags.Concat(e)
		case "import":
			result, e := parseImportBlock(block)
			if result != nil {
				e = e.Concat(appendImport(file.Imports, result))
			}
			diags = diags.Concat(e)
		case "mock":
			result, e := parseMockBlock(block)
			if result != nil {
				e = e.Concat(appendItem(file.Mocks, *result))
			}
			diags = diags.Concat(e)
		case "module":
			result, e := parseV1ModuleImportBlock(block)
			if result != nil {
				e = e.Concat(appendImport(file.Imports, result))
			}
			diags = diags.Concat(e)
		case "param":
			result, e := parseParamBlock(block)
			if result != nil {
				e = e.Concat(appendItem(file.Params, *result))
			}
			diags = diags.Concat(e)
		case "policy":
			result, e := parsePolicyBlock(block)
			if result != nil {
				e = e.Concat(appendItem(file.Policies, *result))
			}
			diags = diags.Concat(e)
		case "sentinel":
			result, e := parseSentinelBlock(block)
			if result != nil {
				if file.SentinelOptions == nil {
					file.SentinelOptions = result
				} else {
					diags = diags.Add(createDuplicateBlockDiag(file.SentinelOptions, block.DefRange))
				}
			}
			diags = append(diags, e...)
		case "test":
			result, e := parseTestBlock(block)
			if result != nil {
				if file.Test == nil {
					file.Test = result
				} else {
					diags = diags.Add(createDuplicateBlockDiag(file.Test, block.DefRange))
				}
			}
			diags = append(diags, e...)
		default:
			diags = diags.Add(&diagnostics.Diagnostic{
				Severity: diagnostics.Error,
				Summary:  "Unexpected block",
				Detail:   fmt.Sprintf("Unknown block called %s", block.Type),
				Range:    convertHclRange(&block.DefRange),
			})
		}
	}

	diags = diags.Concat(ValidateFile(file, sentinelVersion))

	return file, diags
}

func createDuplicateBlockDiag(firstBlock ast.HCLNode, newBlockRange hcl.Range) *diagnostics.Diagnostic {
	return &diagnostics.Diagnostic{
		Severity: diagnostics.Error,
		Summary:  fmt.Sprintf("Duplicate %s block defined", firstBlock.BlockName()),
		Detail:   fmt.Sprintf("%s block has already been defined.", firstBlock.BlockName()),
		Range:    convertHclRange(&newBlockRange),
		Related: &[]diagnostics.RelatedDiagnostic{
			{
				Summary: fmt.Sprintf("First %s block definition.", firstBlock.BlockType()),
				Range:   ast.CloneRange(firstBlock.Range()),
			},
		},
	}
}

func getLabel(block *hcl.Block, index int, defaultValue string) string {
	if index > len(block.Labels) {
		return defaultValue
	}
	return block.Labels[index]
}

func getLabelRange(block *hcl.Block, index int) *position.SourceRange {
	if index > len(block.LabelRanges) {
		return nil
	}
	return hclToSourceRange(&(block.LabelRanges[index]))
}

func getLabelAndRange(block *hcl.Block, index int, defaultValue string) (string, *position.SourceRange) {
	return getLabel(block, index, defaultValue), getLabelRange(block, index)
}
