package position

import "fmt"

type SourceRange struct {
	Filename string    `json:"filename"`
	Start    SourcePos `json:"start"`
	End      SourcePos `json:"end"`
}

type SourcePos struct {
	Line   int `json:"line"`   // Zero based line number
	Column int `json:"column"` // Zero based column
	Byte   int `json:"byte"`
}

func (sr SourceRange) ToString() string {
	return fmt.Sprintf("%d:%d -> %d:%d", sr.Start.Line, sr.Start.Column, sr.End.Line, sr.End.Column)
}

func (sr SourceRange) Clone() SourceRange {
	return SourceRange{
		Filename: sr.Filename,
		Start:    sr.Start.Clone(),
		End:      sr.End.Clone(),
	}
}

func (sr SourcePos) Clone() SourcePos {
	return SourcePos{
		Line:   sr.Line,
		Column: sr.Column,
		Byte:   sr.Byte,
	}
}
