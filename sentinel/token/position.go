package token

type SourcePosition struct {
	Filename string
	Line     int // Line number starting at 0
	Column   int // Column starting at 0
}

func newInvalidSourcePosition() SourcePosition {
	return SourcePosition{Line: -1, Column: -1}
}

func (o SourcePosition) IsValid() bool {
	return o.Line != -1 || o.Column != -1
}

type Pos int
