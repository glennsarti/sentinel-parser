package lexer

// Save the current state
type lexerState struct {
	position     int
	readPosition int
	ch           byte
}

func (l *lexer) getState() lexerState {
	return lexerState{
		position:     l.position,
		readPosition: l.readPosition,
		ch:           l.ch,
	}
}

func (l *lexer) setState(state lexerState) {
	l.position = state.position
	l.readPosition = state.readPosition
	l.ch = state.ch
}
