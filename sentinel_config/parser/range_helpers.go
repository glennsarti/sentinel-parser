package parser

import (
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/hashicorp/hcl/v2"
)

func hclToSourceRange(src *hcl.Range) *position.SourceRange {
	if src == nil {
		return nil
	}

	// HCL Ranges start at Line/Column 1, whereas we start at Line/Column 0
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

func hclToExplicitSourceRange(from, to hcl.Range) *position.SourceRange {
	// HCL Ranges start at Line/Column 1, whereas we start at Line/Column 0
	return &position.SourceRange{
		Filename: from.Filename,
		Start: position.SourcePos{
			Byte:   from.Start.Byte,
			Column: from.Start.Column - 1,
			Line:   from.Start.Line - 1,
		},
		End: position.SourcePos{
			Byte:   to.End.Byte,
			Column: to.End.Column - 1,
			Line:   to.End.Line - 1,
		},
	}
}
