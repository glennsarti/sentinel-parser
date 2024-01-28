package parser

import (
	"fmt"
	"strings"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/sentinel"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
	"github.com/glennsarti/sentinel-parser/sentinel/tracing"
)

type parserState struct {
	lex    sentinel.Lexer
	tracer tracing.ParsingTracer

	parserErrors diagnostics.Diagnostics

	filename string
	locator  token.Locator

	curToken           token.Token
	curTokenStartPos   int
	curTokenEndPos     int
	curBatchedComments *batchedComments

	peekToken           token.Token
	peekTokenStartPos   int
	peekTokenEndPos     int
	peekBatchedComments *batchedComments

	leadingComments ast.Comments
}

func (p *parser) newParserState(
	lex sentinel.Lexer,
	locator token.Locator,
	tracer tracing.ParsingTracer,
) (*parserState, error) {

	if p.sentinelVersion != lex.SentinelVersion() {
		return nil, fmt.Errorf("lexer version %s is not the same as parser version %s", lex.SentinelVersion(), p.sentinelVersion)
	}

	state := &parserState{
		lex:     lex,
		locator: locator,
		tracer:  tracer,

		// TODO: What should the default filename be
		filename: "<filename>",

		curToken:           token.Token{},
		curTokenStartPos:   -1,
		curTokenEndPos:     -1,
		curBatchedComments: nil,

		peekToken:           token.Token{},
		peekTokenStartPos:   -1,
		peekTokenEndPos:     -1,
		peekBatchedComments: nil,

		parserErrors: diagnostics.EmptyDiags(),
		leadingComments: ast.Comments{
			List: make([]*ast.Comment, 0),
		},
	}

	// Populate the current and peeked tokens
	state.nextToken()
	state.nextToken()

	return state, nil
}

// Tracing
func (ps *parserState) traceMethodStart(method string) {
	if ps.tracer != nil {
		ps.tracer.PushMethod(method, fmt.Sprintf("[Pos:%d] %q", ps.curTokenStartPos, ps.curToken.Literal))
	}
}

func (ps *parserState) traceMethodEnd() {
	if ps.tracer != nil {
		ps.tracer.PopMethod(fmt.Sprintf("[Pos:%d] %q", ps.curTokenStartPos, ps.curToken.Literal))
	}
}

func (ps *parserState) traceMessage(message string) {
	if ps.tracer != nil {
		ps.tracer.Trace(message)
	}
}

func (ps *parserState) addExpectedParserError(expected token.TokenType, startPos, endPos int) {
	msg := fmt.Sprintf("Expected '%s', found '%s'", expected, ps.curToken.Type)
	ps.traceMessage(fmt.Sprintf("Parse Error: %s", msg))
	r := ps.newSourceRangeIntInt(startPos, endPos)
	ps.parserErrors = ps.parserErrors.Add(&diagnostics.Diagnostic{
		Severity: diagnostics.Error,
		Summary:  "Parser error",
		Detail:   msg,
		Range:    &r,
	})
}

func (ps *parserState) addExpectedStringParserError(expected string) {
	msg := fmt.Sprintf("Expected %s, found '%s'", expected, ps.curToken.Type)
	ps.traceMessage(fmt.Sprintf("Parse Error: %s", msg))
	r := ps.curTokenSourceRange()
	ps.parserErrors = ps.parserErrors.Add(&diagnostics.Diagnostic{
		Severity: diagnostics.Error,
		Summary:  "Parser error",
		Detail:   msg,
		Range:    &r,
	})
}

func (ps *parserState) newComment(start, end int, literal string) *ast.Comment {
	comment := ast.Comment{
		NodePos: ps.newSourceRangeIntInt(start, end),
	}

	// Prefix
	if strings.HasPrefix(literal, "#") {
		comment.Prefix = "#"
		comment.PrefixPos = ps.newSourceRangeIntInt(start, start+1)
		literal = literal[1:]
	} else if strings.HasPrefix(literal, "//") {
		comment.Prefix = "//"
		comment.PrefixPos = ps.newSourceRangeIntInt(start, start+2)
		literal = literal[2:]
	} else if strings.HasPrefix(literal, "/*") {
		comment.Prefix = "/*"
		comment.PrefixPos = ps.newSourceRangeIntInt(start, start+2)
		literal = literal[2:]
	} else {
		comment.PrefixPos = ps.newSourceRangeIntInt(token.NoPos, token.NoPos)
	}

	// Trim the comment
	literal = strings.TrimLeft(literal, " \t")
	comment.Text = literal
	comment.TextPos = ps.newSourceRangeIntInt(end-len(literal), end)

	return &comment
}

func (ps *parserState) getTrailingComments(tokenPos int) *ast.Comments {
	if ps.curBatchedComments == nil {
		return nil
	}
	curPos := ps.lex.Locator().PositionOf(tokenPos)
	if !curPos.IsValid() || curPos.Line == 0 {
		return nil
	}
	return ps.curBatchedComments.TrailingComments(curPos.Line - 1)
}
