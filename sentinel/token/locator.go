package token

const NoPos = -1

type Locator interface {
	PositionOf(int) SourcePosition
	AddLine(int)
	Filename() string
}

func NewProgressiveLocator() Locator {
	return &locator{
		offsets: make([]int, 0),
		eofPos:  NoPos,
		static:  false,
	}
}

func NewContentLocator(filename string, content []byte) Locator {
	lines := make([]int, 0)
	for offset, b := range content {
		if b == '\n' {
			lines = append(lines, offset+1)
		}
	}

	return &locator{
		offsets:  lines,
		eofPos:   len(content) - 1,
		static:   true,
		filename: filename,
	}
}

type locator struct {
	offsets  []int
	eofPos   int
	static   bool
	filename string
}

func (l *locator) Filename() string {
	return l.filename
}

func (l *locator) PositionOf(pos int) SourcePosition {
	// EOF is a bit special. _Technically_ EOF is not a byte in the file but there can be
	// cases where we want to reference that we found an error at the end of the file. So
	// we do allow to position the last byte in the file.
	if (l.eofPos != NoPos && pos > (l.eofPos+1)) || pos == NoPos {
		return newInvalidSourcePosition()
	}

	prevLineOffset := 0
	for idx, offset := range l.offsets {
		if pos >= prevLineOffset && pos < offset {
			return SourcePosition{
				Filename: l.filename,
				Line:     idx,
				Column:   int(pos - prevLineOffset),
			}
		}
		prevLineOffset = offset
	}

	return SourcePosition{
		Filename: l.filename,
		Line:     len(l.offsets),
		Column:   int(pos - prevLineOffset),
	}
}

func (l *locator) AddLine(linenum int) {
	if l.static {
		return
	}

	if len(l.offsets) > 0 && linenum <= l.offsets[len(l.offsets)-1] {
		return
	}

	l.offsets = append(l.offsets, linenum)
}
