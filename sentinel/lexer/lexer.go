package lexer

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

type lexer struct {
	input           string
	inputLen        int
	position        int
	readPosition    int
	ch              byte
	locator         token.Locator
	implySemicolon  bool
	sentinelVersion string
}

func New(sentinelVersion, input string) (sentinel.Lexer, error) {
	ok, ver := features.ValidateSentinelVersion(sentinelVersion)
	if !ok {
		return nil, fmt.Errorf("%s is an invalid sentinel version", sentinelVersion)
	}

	l := &lexer{input: input,
		locator:         token.NewProgressiveLocator(),
		sentinelVersion: ver,
		inputLen:        len(input),
	}
	l.readChar()
	return l, nil
}

func (l *lexer) Locator() token.Locator { return l.locator }

func (l *lexer) SentinelVersion() string { return l.sentinelVersion }

func (l *lexer) withTrailingEquals(plainTok, equalsTok token.TokenType) token.Token {
	if l.peekChar() == '=' {
		ch := l.ch
		l.readChar()
		literal := string(ch) + string(l.ch)
		return token.Token{Type: equalsTok, Literal: literal}
	} else {
		return newToken(plainTok, l.ch)
	}
}

func (l *lexer) NextToken() (int, int, token.Token) {
	var tok token.Token

	if l.implySemicolon {
		l.skipWhitespace(false)
		if l.ch == '\n' {
			return l.impliedSemicolon()
		}
	} else {
		l.skipWhitespace(true)
	}
	startPos := l.position

	wasImpliedSemicolon := l.implySemicolon
	l.implySemicolon = false

	switch {
	case '1' <= l.ch && l.ch <= '9':
		l.implySemicolon = true
		return l.readNumber()
	case isLetter(l.ch):
		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdent(l.sentinelVersion, tok.Literal)

		tokStart := startPos
		tokEnd := l.position
		if tok.Type == token.IS {
			tokStart, tokEnd, tok = l.readAfterIsToken(startPos)
		} else if tok.Type == token.NOTSTR {
			tokStart, tokEnd, tok = l.readNotToken(tokStart, tokEnd, tok)
		}

		// Only certain tokens require an implied semicolon
		switch tok.Type {
		case token.IDENT, token.BREAK, token.CONTINUE,
			token.ISEMPTY, token.ISNOTEMPTY,
			token.ISDEFINED, token.ISNOTDEFINED:
			l.implySemicolon = true
		default:
			l.implySemicolon = false
		}

		return tokStart, tokEnd, tok

	default:

		switch l.ch {
		case 0:
			return l.inputLen, l.inputLen, token.Token{Type: token.EOF}

		case '+':
			tok = l.withTrailingEquals(token.ADD, token.ADD_ASSIGN)
		case '-':
			tok = l.withTrailingEquals(token.SUB, token.SUB_ASSIGN)
		case '*':
			tok = l.withTrailingEquals(token.MUL, token.MUL_ASSIGN)
		case '/':

			switch l.peekChar() {
			case '=':
				ch := l.ch
				l.readChar()
				literal := string(ch) + string(l.ch)
				tok = token.Token{Type: token.QUO_ASSIGN, Literal: literal}
			case '/':
				// "// ..." comments always end in a new line
				if wasImpliedSemicolon {
					return l.impliedSemicolon()
				}

				tok.Literal = l.readUntilNewLine()
				tok.Type = token.COMMENT
			case '*':
				return l.readMultilineComment(wasImpliedSemicolon)
			default:
				tok = newToken(token.QUO, l.ch)
			}
		case '%':
			tok = l.withTrailingEquals(token.REM, token.REM_ASSIGN)

		case '=':
			tok = l.withTrailingEquals(token.ASSIGN, token.EQL)
		case '<':
			tok = l.withTrailingEquals(token.LSS, token.LEQ)
		case '>':
			tok = l.withTrailingEquals(token.GTR, token.GEQ)
		case '!':
			tok = l.withTrailingEquals(token.NOT, token.NEQ)

		case '.':
			if isDigit(l.peekChar()) {
				l.implySemicolon = true
				return l.readNumber()
			}
			tok = newToken(token.PERIOD, l.ch)
		case ';':
			l.implySemicolon = false
			tok = newToken(token.SEMICOLON, l.ch)
		case ':':
			tok = newToken(token.COLON, l.ch)
		case ',':
			tok = newToken(token.COMMA, l.ch)
		case '{':
			l.implySemicolon = false
			tok = newToken(token.LBRACE, l.ch)
		case '}':
			l.implySemicolon = true
			tok = newToken(token.RBRACE, l.ch)
		case '(':
			l.implySemicolon = false
			tok = newToken(token.LPAREN, l.ch)
		case ')':
			l.implySemicolon = true
			tok = newToken(token.RPAREN, l.ch)
		case '"':
			l.implySemicolon = true
			tok.Type = token.STRING
			tok.Literal = `"` + l.readString() + `"`
		case '[':
			l.implySemicolon = false
			tok = newToken(token.LBRACK, l.ch)
		case ']':
			l.implySemicolon = true
			tok = newToken(token.RBRACK, l.ch)

		case '#':
			// "... # ..." comments always end in a new line
			if wasImpliedSemicolon {
				return l.impliedSemicolon()
			}
			tok.Literal = l.readUntilNewLine()
			tok.Type = token.COMMENT

		case '0':
			l.implySemicolon = true
			c := l.peekChar()
			if c == 'x' || c == 'X' {
				l.readChar() // Read the x|X
				l.readChar() // and then read the next char
				return l.readHexNumber(startPos)
			}
			return l.readNumber()
		}
	}

	l.readChar()
	return startPos, l.position, tok
}

func (l *lexer) impliedSemicolon() (int, int, token.Token) {
	l.implySemicolon = false
	return l.position, l.position + 1, token.Token{
		Type:    token.SEMICOLON,
		Literal: "\n",
	}
}

// Language Spec for Numbers
//
// letter        = unicode_letter | "_" .
// decimal_digit = "0" … "9" .
// octal_digit   = "0" … "7" .
// hex_digit     = "0" … "9" | "A" … "F" | "a" … "f" .
//
// int_lit     = decimal_lit | octal_lit | hex_lit .
// decimal_lit = ( "1" … "9" ) { decimal_digit } .
// octal_lit   = "0" { octal_digit } .
// hex_lit     = "0" ( "x" | "X" ) hex_digit { hex_digit } .
//
// float_lit = decimals "." [ decimals ] [ exponent ] |
//
//	decimals exponent |
//	"." decimals [ exponent ] .
//
// decimals  = decimal_digit { decimal_digit } .
// exponent  = ( "e" | "E" ) [ "+" | "-" ] decimals .
func (l *lexer) readNumber() (int, int, token.Token) {
	startPos := l.position
	startDigit := l.ch
	preDecimal := ""

	if isDigit(l.ch) {
		preDecimal = l.readDigits()
	}

	if l.ch == '.' {
		if preDecimal != "" {
			l.readChar()
			// FLOAT: decimals "." [ decimals ] [ exponent ]
			//                     ^

			if isDigit(l.ch) {
				l.readDigits()
			}

			// FLOAT: decimals "." [ decimals ] [ exponent ]
			//                                  ^
			if (l.ch == 'e' || l.ch == 'E') && !l.readExponent() {
				return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
			}
			return l.newPositionedToken(startPos, l.position, token.FLOAT)
		} else {
			l.readChar()
			// FLOAT: "." decimals [ exponent ] .
			//            ^
			if !isDigit(l.ch) {
				return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
			}
			l.readDigits()

			// FLOAT: "." decimals [ exponent ] .
			//                     ^
			if (l.ch == 'e' || l.ch == 'E') && !l.readExponent() {
				return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
			}
			return l.newPositionedToken(startPos, l.position, token.FLOAT)
		}
	} else if l.ch == 'e' || l.ch == 'E' {
		// FLOAT: decimals exponent |
		//                 ^
		if preDecimal == "" {
			return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
		}
		if !l.readExponent() {
			return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
		}
		return l.newPositionedToken(startPos, l.position, token.FLOAT)
	} else if preDecimal == "" {
		// ILLEGAL: a
		//          ^
		return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
	}

	// Base 8 integers don't start with '0'
	if startDigit == '0' {
		for _, c := range preDecimal {
			if c == '8' || c == '9' {
				return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
			}
		}
	}

	return l.newPositionedToken(startPos, l.position, token.INT)
}

// exponent = ( "e" | "E" ) [ "+" | "-" ] decimals .
func (l *lexer) readExponent() bool {
	if l.ch != 'e' && l.ch != 'E' {
		return false
	}
	l.readChar()
	if l.ch == '+' || l.ch == '-' {
		l.readChar()
	}

	// Need at least one digit
	if !isDigit(l.ch) {
		return false
	}

	l.readDigits()

	return true
}

func (l *lexer) skipWhitespace(skipNewLine bool) {
	for l.ch == ' ' ||
		l.ch == '\t' ||
		(skipNewLine && l.ch == '\n') ||
		l.ch == '\r' {
		l.readChar()
	}
}

func (l *lexer) readChar() {
	if l.readPosition >= l.inputLen {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
		l.position = l.readPosition
		l.readPosition += 1

		if l.ch == '\n' {
			l.locator.AddLine(l.readPosition)
		}
	}
}

func (l *lexer) readMultilineComment(wasImpliedSemicolon bool) (int, int, token.Token) {
	startPos := l.position

	beforeCommentState := l.getState()

	if l.ch != '/' {
		return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
	}

	l.readChar()
	if l.ch != '*' {
		return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
	}

	l.readChar()
	for l.ch != 0 {
		if l.ch == '\n' && wasImpliedSemicolon {
			// `/* ... */` comments which span lines will inject an implied semicolon
			l.setState(beforeCommentState)
			return l.impliedSemicolon()
		} else if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()

			// `/* ... */` comments and implied semicolons are "complicated"
			// so delegate the logic to peekTrailingCommentsHaveNewLine()
			if wasImpliedSemicolon && l.peekTrailingCommentsHaveNewLine() {
				l.setState(beforeCommentState)
				return l.impliedSemicolon()
			}

			return l.newPositionedToken(startPos, l.position, token.COMMENT)
		}
		l.readChar()
	}
	return l.newPositionedToken(startPos, l.position, token.ILLEGAL)
}

func (l *lexer) peekTrailingCommentsHaveNewLine() bool {
	// This is a little complicated but ... this function tries to find out if the trailing
	// character after all consecutive comments have been consumed have a new line either
	// in the comments themselves, or directly after it.
	// Examples;
	//
	// Yes:
	// - /* Ignore */ \n                         <-- Trailing `\n`
	// - /* Ignore */ /* Ignore \n */            <-- `\n` mid-comment
	// - /* Ignore */ /* Ignore */ // Comment    <-- `//` always end in a newline
	// - /* Ignore */ /* Ignore */ # Comment     <-- `#` always end in a newline
	// No:
	// - /* Ignore */ /* Ignore */ statement     <-- Doesn't end in `\n`
	// - /* Ignore */ /* Ignore */;              <-- Doesn't end in `\n`
	thisState := l.getState()
	// Always revert the lexer state because we're peeking.
	defer l.setState(thisState)

	if l.ch == 0 {
		return false
	}

	for {
		l.skipWhitespace(false)

		switch l.ch {
		case 0:
			// EOF
			return false
		case '\n':
			// Newline after comment
			return true
		case '#':
			// `#` comments always end in a newline
			return true

		case '/':
			// '/....' characters could be a number of different sequences
			l.readChar()
			switch l.ch {
			case '/':
				// `//` comments always end in a newline
				return true
			case '*':
				l.readChar()
				// We've hit another mutlitline comment so read until
				// the end of the comment OR a newline mid-comment (OR reach EOF)
				for l.ch != 0 {
					l.readChar()
					switch l.ch {
					case 0:
						// EOF
						return false
					case '\n':
						// New line mid-comment
						return true
					case '*':
						l.readChar()
						if l.ch == '/' {
							// End of comment marker `*/`
							break
						}
					}
				}
			default:
				// Anything else isn't a newline, so return false.
				return false
			}
		default:
			// Anything else isn't a newline, so return false.
			return false
		}
	}
}

func (l *lexer) peekChar() byte {
	if l.readPosition >= l.inputLen {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) readWord() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) readDigits() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) readHexNumber(startPos int) (int, int, token.Token) {
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return l.newPositionedToken(startPos, l.position, token.INT)
}

func (l *lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *lexer) readUntilNewLine() string {
	position := l.position
	for {
		ch := l.peekChar()
		if ch == 0 || ch == 13 || ch == 10 {
			return l.input[position : l.position+1]
		}
		l.readChar()
	}
}

func (l *lexer) readAfterIsToken(isPos int) (int, int, token.Token) {
	state := l.getState()

	// Read the next identifer after "is"
	l.skipWhitespace(true)
	nextToken := token.LookupIdent(l.sentinelVersion, l.readWord())

	if nextToken == token.NOTSTR {
		return l.readIsNotToken(isPos)
	}

	switch nextToken {
	case token.EMPTY:
		return l.newPositionedToken(isPos, l.position, token.ISEMPTY)
	case token.DEFINED:
		//		if l.supportIsDefined {
		return l.newPositionedToken(isPos, l.position, token.ISDEFINED)
		//		}
	}
	l.setState(state)
	return l.newPositionedToken(isPos, state.position, token.IS)
}

func (l *lexer) readIsNotToken(isNotPos int) (int, int, token.Token) {
	state := l.getState()

	// Read the next identifer after "is not"
	l.skipWhitespace(true)
	nextToken := token.LookupIdent(l.sentinelVersion, l.readWord())

	switch nextToken {
	case token.EMPTY:
		return l.newPositionedToken(isNotPos, l.position, token.ISNOTEMPTY)
	case token.DEFINED:
		//		if l.supportIsDefined {
		return l.newPositionedToken(isNotPos, l.position, token.ISNOTDEFINED)
		//		}
	}
	l.setState(state)
	return l.newPositionedToken(isNotPos, state.position, token.ISNOT)
}

func (l *lexer) readNotToken(tokStart, tokEnd int, tok token.Token) (int, int, token.Token) {
	state := l.getState()

	// Read the next identifer after "is"
	l.skipWhitespace(true)
	nextToken := token.LookupIdent(l.sentinelVersion, l.readWord())

	switch nextToken {
	case token.CONTAINS:
		return l.newPositionedToken(tokStart, l.position, token.NOTCONTAINS)
	case token.IN:
		return l.newPositionedToken(tokStart, l.position, token.NOTIN)
	case token.MATCHES:
		return l.newPositionedToken(tokStart, l.position, token.NOTMATCHES)
	}
	l.setState(state)
	return tokStart, tokEnd, tok
}

func (l *lexer) newPositionedToken(start, end int, tokenType token.TokenType) (int, int, token.Token) {
	return start, end, token.Token{Type: tokenType, Literal: string(l.input[start:end])}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ('A' <= ch && ch <= 'F') || ('a' <= ch && ch <= 'f')
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
