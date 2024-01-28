package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/hashicorp/hcl/v2"
)

func errorToDiagnostics(err error) diagnostics.Diagnostics {
	return diagnostics.Diagnostics{&diagnostics.Diagnostic{
		Severity: diagnostics.Error,
		Summary:  fmt.Sprintf("%s", err),
	},
	}
}

func convertHclDiagnostics(src hcl.Diagnostics) diagnostics.Diagnostics {
	diags := make(diagnostics.Diagnostics, len(src))

	for idx := range src {
		diags[idx] = convertHclDiagnostic(src[idx])
	}

	return diags
}

func convertHclDiagnostic(src *hcl.Diagnostic) *diagnostics.Diagnostic {
	if src == nil {
		return nil
	}
	return &diagnostics.Diagnostic{
		Detail:   src.Detail,
		Severity: convertHclSeverity(src.Severity),
		Range:    convertHclRange(src.Subject),
		Summary:  src.Summary,
	}
}

func convertHclSeverity(sev hcl.DiagnosticSeverity) diagnostics.SeverityLevel {
	switch sev {
	case hcl.DiagError:
		return diagnostics.Error
	case hcl.DiagWarning:
		return diagnostics.Warning
	default:
		return diagnostics.Unknown
	}
}

func convertHclRange(src *hcl.Range) *position.SourceRange {
	if src == nil {
		return nil
	}

	// HCL Diagnostics start at Line/Column 1, whereas we start at Line/Column 0
	return &position.SourceRange{
		Filename: src.Filename,
		Start: position.SourcePos{
			Byte:   src.Start.Byte,
			Column: src.Start.Column - 1,
			Line:   src.Start.Line - 1,
		},
		End: position.SourcePos{
			Byte:   src.End.Byte,
			Column: src.End.Column - 1,
			Line:   src.End.Line - 1,
		},
	}
}

func concatHclDiagnostics(src diagnostics.Diagnostics, d hcl.Diagnostics) diagnostics.Diagnostics {
	return src.Concat(convertHclDiagnostics(d))
}
