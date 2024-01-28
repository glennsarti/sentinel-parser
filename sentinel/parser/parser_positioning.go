package parser

import (
	"github.com/glennsarti/sentinel-parser/position"
)

func (ps *parserState) newInvalidSourceRange() position.SourceRange {
	return position.SourceRange{
		Filename: ps.locator.Filename(),
		Start: position.SourcePos{
			Byte:   -1,
			Column: -1,
			Line:   -1,
		},
		End: position.SourcePos{
			Byte:   -1,
			Column: -1,
			Line:   -1,
		},
	}
}

func (ps *parserState) newSourceRangeIntInt(start, end int) position.SourceRange {
	startPos := ps.locator.PositionOf(start)
	endPos := ps.locator.PositionOf(end)
	return position.SourceRange{
		Filename: ps.locator.Filename(),
		Start: position.SourcePos{
			Byte:   start,
			Column: startPos.Column,
			Line:   startPos.Line,
		},
		End: position.SourcePos{
			Byte:   end,
			Column: endPos.Column,
			Line:   endPos.Line,
		},
	}
}

func (ps *parserState) newSourceRangeIntPos(start int, endPos position.SourcePos) position.SourceRange {
	startPos := ps.locator.PositionOf(start)
	return position.SourceRange{
		Filename: ps.locator.Filename(),
		Start: position.SourcePos{
			Byte:   start,
			Column: startPos.Column,
			Line:   startPos.Line,
		},
		End: endPos.Clone(),
	}
}

func (ps *parserState) newSourceRangePosInt(startPos position.SourcePos, end int) position.SourceRange {
	endPos := ps.locator.PositionOf(end)
	return position.SourceRange{
		Filename: ps.locator.Filename(),
		Start:    startPos.Clone(),
		End: position.SourcePos{
			Byte:   end,
			Column: endPos.Column,
			Line:   endPos.Line,
		},
	}
}

func (ps *parserState) newSourceRangePosPos(startPos, endPos position.SourcePos) position.SourceRange {
	return position.SourceRange{
		Filename: ps.locator.Filename(),
		Start:    startPos.Clone(),
		End:      endPos.Clone(),
	}
}

func (ps *parserState) sourcePosOf(offset int) position.SourcePos {
	pos := ps.locator.PositionOf(offset)
	return position.SourcePos{
		Byte:   offset,
		Column: pos.Column,
		Line:   pos.Line,
	}
}
