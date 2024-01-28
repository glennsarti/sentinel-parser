package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

// Token Helpers

type tokenLocation struct {
	Start int
	End   int
}

// Expects the token to be a particular type. If it is not then we increment
// the lexer, and raise an error
func (ps *parserState) expectTokenIs(tt token.TokenType) bool {
	if ps.curTokenIs(tt) {
		return true
	}

	r := ps.curTokenSourceRange()
	ps.parserErrors = ps.parserErrors.Add(&diagnostics.Diagnostic{
		Severity: diagnostics.Error,
		Summary:  "Parser error",
		Detail:   fmt.Sprintf("Expected '%s', but found '%s'", tt, ps.curToken.Type),
		Range:    &r,
	})

	ps.nextToken()
	return false
}

func (ps *parserState) consumeEndOfStatements() {
	for ps.curTokenIs(token.SEMICOLON) {
		ps.nextToken()
	}
}

func (ps *parserState) consumePeekEndOfStatements() {
	for ps.peekTokenIs(token.SEMICOLON) {
		ps.nextToken()
	}
}

func (ps *parserState) nextToken() {
	ps.curToken = ps.peekToken
	ps.curTokenStartPos = ps.peekTokenStartPos
	ps.curTokenEndPos = ps.peekTokenEndPos
	ps.curBatchedComments = ps.peekBatchedComments

	ps.peekTokenStartPos,
		ps.peekTokenEndPos,
		ps.peekToken = ps.lex.NextToken()

	// Consume any comments if ...
	// ... If it's not a comment, return now
	if ps.peekToken.Type != token.COMMENT {
		return
	}

	ps.peekBatchedComments = nil
	// ... We should only batch comments after the current token line because
	// we're only interesting in comments at the beginning of a line
	afterTokenLine := ps.lex.Locator().PositionOf(ps.curTokenEndPos).Line + 1

	for {
		if ps.peekToken.Type != token.COMMENT {
			return
		}
		if ps.peekBatchedComments == nil {
			// Check if we've started on a new line first
			peekTokenLine := ps.lex.Locator().PositionOf(ps.peekTokenStartPos).Line
			if peekTokenLine >= afterTokenLine {
				ps.peekBatchedComments = &batchedComments{Comments: []*ast.Comment{
					ps.newComment(ps.peekTokenStartPos, ps.peekTokenEndPos, ps.peekToken.Literal),
				}}
			}
		} else {
			ps.peekBatchedComments.Comments = append(ps.peekBatchedComments.Comments,
				ps.newComment(ps.peekTokenStartPos, ps.peekTokenEndPos, ps.peekToken.Literal),
			)
		}

		ps.peekTokenStartPos,
			ps.peekTokenEndPos,
			ps.peekToken = ps.lex.NextToken()
	}
}

func (ps *parserState) curTokenLocation() tokenLocation {
	return tokenLocation{
		Start: ps.curTokenStartPos,
		End:   ps.curTokenEndPos,
	}
}

func (ps *parserState) curTokenSourceRange() position.SourceRange {
	return ps.newSourceRangeIntInt(ps.curTokenStartPos, ps.curTokenEndPos)
}

func (ps *parserState) curTokenIs(tt token.TokenType) bool {
	return ps.curToken.Type == tt
}

func (ps *parserState) peekTokenIs(tt token.TokenType) bool {
	return ps.peekToken.Type == tt
}

// Expects end of statement markers ";". Returns true if found
// Consumes the semi-colon if it is found
func (ps *parserState) expectPeekEndOfStmt() bool {
	// Semicolons are optional for closing ')' or '}'
	if ps.peekTokenIs(token.RPAREN) || ps.peekTokenIs(token.RBRACE) {
		return true
	}
	if ps.peekTokenIs(token.SEMICOLON) {
		ps.nextToken()
		return true
	}

	ps.nextToken()
	ps.addExpectedParserError(token.SEMICOLON, ps.curTokenStartPos, ps.curTokenEndPos)
	return false
}

// Expects the token to be a quantifier op token. If it is not then we raise an error
func (ps *parserState) expectTokenIsQuantOp() bool {
	switch ps.curToken.Type {
	case token.ALL,
		token.ANY,
		token.FILTER,
		token.MAP:
		return true
	default:
		r := ps.curTokenSourceRange()
		ps.parserErrors = ps.parserErrors.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Parser error",
			Detail:   fmt.Sprintf("Expected one of 'any', 'all', 'filter' or 'map', but found '%s'", ps.curToken.Type),
			Range:    &r,
		})
	}

	return false
}

// Consume trailing semicolon if it exists
// Typically happens after '}' or ')'
// e.g.   `... };` or `... }\n`
func (ps *parserState) consumePeekSemicolon() {
	if ps.peekTokenIs(token.SEMICOLON) {
		ps.nextToken()
	}
}

// nolint
// Only needed for debugging
// func (ps *parserState) debugTokens(title string) {
// 	ps.traceMessage("--- " + title)
// 	ps.traceMessage(fmt.Sprintf(
// 		"Current Token (%d) %s -> %q",
// 		ps.curTokenStartPos,
// 		strings.ReplaceAll(string(ps.curToken.Type), "\n", "\\n"),
// 		strings.ReplaceAll(string(ps.curToken.Literal), "\n", "\\n"),
// 	))
// 	ps.traceMessage(fmt.Sprintf(
// 		"Peek Token (%d) %s -> %q",
// 		ps.peekTokenStartPos,
// 		strings.ReplaceAll(string(ps.peekToken.Type), "\n", "\\n"),
// 		strings.ReplaceAll(string(ps.peekToken.Literal), "\n", "\\n"),
// 	))
// 	ps.traceMessage("---")
// }
