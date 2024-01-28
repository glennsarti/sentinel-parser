package parser

import (
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

type (
	unaryParseFunc  func(*parserState) ast.Expression
	binaryParseFunc func(ast.Expression, *parserState) ast.Expression
	suffixParseFunc func(ast.Expression, *parserState) ast.Expression
)

func (p *parser) registerUnaryFunc(tokenType token.TokenType, fn unaryParseFunc) {
	p.unaryParseFuncs[tokenType] = fn
}

func (p *parser) registerBinaryFunc(tokenType token.TokenType, fn binaryParseFunc) {
	p.binaryParseFuncs[tokenType] = fn
}

func (p *parser) registerSuffixFunc(tokenType token.TokenType, fn suffixParseFunc) {
	p.suffixParseFuncs[tokenType] = fn
}
