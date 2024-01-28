package parser

import (
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

// Operator Precedence
// This should be the inverse list of https://developer.hashicorp.com/sentinel/docs/language/spec#operator-precedence
const (
	_ int = iota
	PREC_LOWEST
	PREC_BOOL_OR    // or xor
	PREC_BOOL_AND   // and
	PREC_COMPARISON // == != < <= > >= "is" "is not" "matches" "contains" "in"
	PREC_ELSE       // else
	PREC_ADDITION   // + -
	PREC_MULTIPLY   // * / %
	PREC_UNARY      // Unary has the highest priority
)

var precedences = map[token.TokenType]int{
	// or  xor
	token.LOR:  PREC_BOOL_OR,
	token.LXOR: PREC_BOOL_OR,
	// and
	token.LAND: PREC_BOOL_AND,
	// comparison
	token.EQL:         PREC_COMPARISON,
	token.NEQ:         PREC_COMPARISON,
	token.LSS:         PREC_COMPARISON,
	token.LEQ:         PREC_COMPARISON,
	token.GTR:         PREC_COMPARISON,
	token.GEQ:         PREC_COMPARISON,
	token.IS:          PREC_COMPARISON,
	token.ISNOT:       PREC_COMPARISON,
	token.MATCHES:     PREC_COMPARISON,
	token.NOTMATCHES:  PREC_COMPARISON,
	token.CONTAINS:    PREC_COMPARISON,
	token.NOTCONTAINS: PREC_COMPARISON,
	token.IN:          PREC_COMPARISON,
	token.NOTIN:       PREC_COMPARISON,
	// else
	token.ELSE: PREC_ELSE,
	// + -
	token.ADD: PREC_ADDITION,
	token.SUB: PREC_ADDITION,
	// * / %
	token.MUL: PREC_MULTIPLY,
	token.QUO: PREC_MULTIPLY,
	token.REM: PREC_MULTIPLY,
}

func (ps *parserState) peekPrecedence() int {
	if p, ok := precedences[ps.peekToken.Type]; ok {
		return p
	}

	return PREC_LOWEST
}

func (ps *parserState) curPrecedence() int {
	if p, ok := precedences[ps.curToken.Type]; ok {
		return p
	}

	return PREC_LOWEST
}
