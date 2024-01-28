package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

// Expression = UnaryExpr | Expression binary_op Expression .
// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
//
// Operand     = Literal | OperandName | MethodExpr | "(" Expression ")" .
// Literal     = BasicLit | CompositeLit | FunctionLit | RuleLit .
// BasicLit    = int_lit | float_lit | string_lit .
// OperandName = identifier | QualifiedIdent.
//
// PrimaryExpr =
//
//	Operand |
//	PrimaryExpr Selector |
//	PrimaryExpr Index |
//	PrimaryExpr Slice |
//	PrimaryExpr Arguments .
//
// Selector  = "." identifier .
// Index     = "[" Expression "]" .
// Slice     = "[" [ Expression ] ":" [ Expression ] "]" | "[" [ Expression ] ":" Expression ":" Expression "]" .
//
// Arguments = "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .
//
// binary_op  = logical_op | set_op | rel_op | add_op | mul_op | else_op .
// logical_op = "and" | "or" | "xor" .
// set_op     = ["not"] ( "contains" | "in" ).
// rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" | "is" | "is not" | "matches" | "not matches" .
// add_op     = "+" | "-" .
// mul_op     = "*" | "/" | "%" .
// else_op    = "else" .
// unary_op   = "+" | "-" | "!" | "not" | empty_op | defined_op .
// empty_op   = "is empty" | "is not empty" .
// defined_op = "is defined" | "is not defined" .
func (p *parser) parseExpression(precedence int, state *parserState) ast.Expression {
	state.traceMethodStart("parseExpression")
	defer state.traceMethodEnd()

	unaryFunc := p.unaryParseFuncs[state.curToken.Type]
	if unaryFunc == nil {
		r := state.curTokenSourceRange()
		state.parserErrors = state.parserErrors.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Parsing error",
			Detail:   fmt.Sprintf("No unary parser of '%s'", state.curToken.Type),
			Range:    &r,
		})
		return &ast.BadExpression{NodePos: r}
	}
	leftExp := p.parseExpressionSuffixes(unaryFunc(state), state)

	for !state.peekTokenIs(token.SEMICOLON) && precedence < state.peekPrecedence() {

		binaryFunc := p.binaryParseFuncs[state.peekToken.Type]
		if binaryFunc == nil {
			state.traceMessage(fmt.Sprintf("Missing binary for '%s'", state.peekToken.Type))
			return leftExp
		}
		state.traceMessage(fmt.Sprintf("Found binary for '%s'", state.peekToken.Type))

		state.nextToken()

		leftExp = p.parseExpressionSuffixes(binaryFunc(leftExp, state), state)
	}

	return leftExp
}

func (p *parser) parseBinaryExpr(left ast.Expression, state *parserState) ast.Expression {
	state.traceMethodStart("parseBinaryExpr")
	defer state.traceMethodEnd()

	expr := &ast.BinaryExpression{
		NodePos:  left.Position().Clone(),
		LeftExpr: left,
		Op:       state.curToken.Type,
		OpPos:    state.curTokenSourceRange(),
	}

	precedence := state.curPrecedence()
	state.nextToken()
	expr.RightExpr = p.parseExpression(precedence, state)
	expr.NodePos.End = expr.RightExpr.Position().End

	return expr
}

func (p *parser) parseCallExpression(callee ast.Expression, state *parserState) ast.Expression {
	state.traceMethodStart("parseCallExpr")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "("
	if !state.expectTokenIs(token.LPAREN) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	lpos := state.curTokenSourceRange()

	items := p.parseExpressionList(token.RPAREN, state)

	// ")"
	if !state.expectTokenIs(token.RPAREN) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rpos := state.curTokenSourceRange()

	return &ast.CallExpression{
		NodePos: state.newSourceRangePosPos(callee.Position().Start, rpos.End),

		Callee:     callee,
		LeftParen:  lpos,
		Args:       items,
		RightParen: rpos,
	}
}

func (p *parser) parseExpressionList(endToken token.TokenType, state *parserState) []ast.Expression {
	state.traceMethodStart("parseExpressionList")
	defer state.traceMethodEnd()

	list := []ast.Expression{}

	if state.peekTokenIs(endToken) {
		state.nextToken()
		return list
	}

	state.nextToken()
	list = append(list, p.parseExpression(PREC_LOWEST, state))

	// Expression lists can terminate with
	// [ <Expr> ]
	// OR
	// [
	//   ...
	//   <Expr>,  <--- Mandatory trailing comma on multiline lists
	// ]
	for state.peekTokenIs(token.COMMA) {
		state.nextToken()

		if state.peekTokenIs(endToken) {
			// Terminate if this is a trailing comma; "," "<endToken>"
			break
		}
		state.nextToken()
		list = append(list, p.parseExpression(PREC_LOWEST, state))
	}
	state.nextToken()

	state.expectTokenIs(endToken)
	return list
}

func (p *parser) parseExpressionSuffixes(value ast.Expression, state *parserState) ast.Expression {
	// Process any suffixes
	for {
		suffixFunc := p.suffixParseFuncs[state.peekToken.Type]
		if suffixFunc == nil {
			break
		}
		state.nextToken()
		value = suffixFunc(value, state)
	}
	return value
}

func (p *parser) parseFloatLiteral(state *parserState) ast.Expression {
	state.traceMethodStart("parseFloatLiteral")
	defer state.traceMethodEnd()

	expr := &ast.BasicLit{
		NodePos: state.curTokenSourceRange(),
		Kind:    state.curToken.Type,
		Value:   state.curToken.Literal,
	}
	return expr
}

func (p *parser) parseFuncLiteral(state *parserState) ast.Expression {
	state.traceMethodStart("parseFuncLiteral")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "func"
	if !state.expectTokenIs(token.FUNC) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	fpos := state.curTokenSourceRange()
	state.nextToken()

	fields := p.parseFieldList(state)
	if fields == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	state.nextToken()

	body := p.parseBlockStatement(state)
	if body == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	return &ast.FuncLit{
		NodePos: state.newSourceRangePosPos(fpos.Start, body.Position().End),
		FuncPos: fpos,
		Params:  fields,
		Body:    body,
	}
}

// GroupExpr = "(" Expression ")" .
func (p *parser) parseGroupExpression(state *parserState) ast.Expression {
	state.traceMethodStart("parseGroupExpression")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "("
	if !state.expectTokenIs(token.LPAREN) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	// Rule value
	value := p.parseExpression(PREC_LOWEST, state)
	state.nextToken()
	state.consumeEndOfStatements()

	// ")"
	if !state.expectTokenIs(token.RPAREN) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rpos := state.curTokenSourceRange()

	return &ast.GroupExpression{
		NodePos:    state.newSourceRangePosPos(lpos.Start, rpos.End),
		LeftParen:  lpos,
		Value:      value,
		RightParen: rpos,
	}
}

func (p *parser) parseIdentExpression(state *parserState) ast.Expression {
	state.traceMethodStart("parseIdentExpression")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos
	ident := p.parseIdent(state)
	if ident == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	return ident
}

func (p *parser) parseIndexOrSliceExpression(value ast.Expression, state *parserState) ast.Expression {
	state.traceMethodStart("parseIndexOrSliceExpr")
	defer state.traceMethodEnd()

	startPos := value.Position().Start

	// Index     = "[" Expression "]" .
	// Slice     = "[" [ Expression ] ":" [ Expression ] "]" .

	// "["
	if !state.expectTokenIs(token.LBRACK) {
		return &ast.BadExpression{NodePos: state.newSourceRangePosInt(startPos, state.curTokenEndPos)}
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	var firstExpr ast.Expression = nil
	var secondExpr ast.Expression = nil
	var colonPos position.SourceRange
	var rpos position.SourceRange

	// Empty '[]' is not allowed
	if state.curTokenIs(token.RBRACK) {
		r := state.curTokenSourceRange()
		state.parserErrors = state.parserErrors.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Parsing error",
			Detail:   fmt.Sprintf("Expected an expression, found '%s'", state.curToken.Type),
			Range:    &r,
		})

		return &ast.BadExpression{NodePos: state.newSourceRangePosInt(startPos, state.curTokenEndPos)}
	}

	// Is it an empty low slice?
	if state.curTokenIs(token.COLON) {
		colonPos = state.curTokenSourceRange()
	} else {
		firstExpr = p.parseExpression(PREC_LOWEST, state)

		// Is it a plain index?
		switch state.peekToken.Type {
		case token.RBRACK:
			state.nextToken()

			return &ast.IndexExpression{
				NodePos:    state.newSourceRangePosInt(startPos, state.curTokenEndPos),
				Value:      value,
				LeftBrack:  lpos,
				Index:      firstExpr,
				RightBrack: state.curTokenSourceRange(),
			}
		case token.COLON:
			state.nextToken()
			// Keep on parsing
			colonPos = state.curTokenSourceRange()
		default:
			state.addExpectedParserError(token.RBRACK, state.curTokenStartPos, state.curTokenEndPos)
			return &ast.BadExpression{
				NodePos: state.newSourceRangePosInt(startPos, state.curTokenEndPos),
			}
		}
	}
	state.nextToken()

	// Is it an empty high slice?
	if !state.curTokenIs(token.RBRACK) {
		secondExpr = p.parseExpression(PREC_LOWEST, state)
		state.nextToken()

		if !state.expectTokenIs(token.RBRACK) {
			return &ast.BadExpression{NodePos: state.newSourceRangePosInt(startPos, state.curTokenEndPos)}
		}
	}
	rpos = state.curTokenSourceRange()

	return &ast.SliceExpression{
		NodePos:    state.newSourceRangePosInt(startPos, state.curTokenEndPos),
		Value:      value,
		LeftBrack:  lpos,
		LowExpr:    firstExpr,
		Colon:      colonPos,
		HighExpr:   secondExpr,
		RightBrack: rpos,
	}
}

func (p *parser) parseIntLiteral(state *parserState) ast.Expression {
	state.traceMethodStart("parseIntLiteral")
	defer state.traceMethodEnd()

	return &ast.BasicLit{
		NodePos: state.curTokenSourceRange(),
		Kind:    state.curToken.Type,
		Value:   state.curToken.Literal,
	}
}

func (p *parser) parseKeyedExpression(state *parserState) ast.Expression {
	state.traceMethodStart("parseKeyedExpression")
	defer state.traceMethodEnd()
	startPos := state.curTokenStartPos

	state.traceMessage("Parsing key")
	key := p.parseExpression(PREC_LOWEST, state)

	if !state.peekTokenIs(token.COLON) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	state.nextToken()
	colon := state.curTokenSourceRange()

	state.nextToken()
	state.traceMessage("Parsing value")
	value := p.parseExpression(PREC_LOWEST, state)

	return &ast.KeyedElementExpression{
		NodePos:  state.newSourceRangePosPos(key.Position().Start, value.Position().End),
		Key:      key,
		ColonPos: colon,
		Value:    value,
	}
}

func (p *parser) parseKeyedExpressionList(endToken token.TokenType, state *parserState) []ast.Expression {
	state.traceMethodStart("parseKeyedExpressionList")
	defer state.traceMethodEnd()

	list := []ast.Expression{}

	if state.peekTokenIs(endToken) {
		state.nextToken()
		return list
	}

	state.nextToken()
	list = append(list, p.parseKeyedExpression(state))

	// Expression lists can terminate with
	// { <Expr> : <Expr> }
	// OR
	// ...
	//   <Expr> : <Expr>,  <--- Mandatory trailing comma on multiline maps
	// }
	for state.peekTokenIs(token.COMMA) {
		state.nextToken()

		if state.peekTokenIs(endToken) {
			// Terminate if this is a trailing comma
			break
		}
		state.nextToken()
		list = append(list, p.parseKeyedExpression(state))
	}
	state.nextToken()

	state.expectTokenIs(endToken)
	return list
}

func (p *parser) parseListLit(state *parserState) ast.Expression {
	state.traceMethodStart("parseListLit")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "["
	if !state.expectTokenIs(token.LBRACK) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	lpos := state.curTokenSourceRange()

	items := p.parseExpressionList(token.RBRACK, state)

	if !state.expectTokenIs(token.RBRACK) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rpos := state.curTokenSourceRange()

	return &ast.ListLit{
		NodePos:    state.newSourceRangePosPos(lpos.Start, rpos.End),
		LeftBrack:  lpos,
		Items:      items,
		RightBrack: rpos,
	}
}

func (p *parser) parseMapLiteral(state *parserState) ast.Expression {
	state.traceMethodStart("parseMapLiteral")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "{"
	if !state.expectTokenIs(token.LBRACE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	lpos := state.curTokenSourceRange()

	items := p.parseKeyedExpressionList(token.RBRACE, state)

	// "}"
	if !state.expectTokenIs(token.RBRACE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rpos := state.curTokenSourceRange()

	return &ast.MapLit{
		NodePos:    state.newSourceRangePosPos(lpos.Start, rpos.End),
		LeftBrace:  lpos,
		Elements:   items,
		RightBrace: rpos,
	}
}

// QuantExpr = QuantOp Expression "as" [ identifier "," ] identifier "{" BoolExpr "}" .
// QuantOp   = "all" | "any" | "filter" | "map" .
func (p *parser) parseQuantExpression(state *parserState) ast.Expression {
	state.traceMethodStart("parseQuantExpression")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// We need a QuantOp first
	if !state.expectTokenIsQuantOp() {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	op := state.curToken.Type
	opPos := state.curTokenSourceRange()
	state.nextToken()

	// Value to quantify
	value := p.parseExpression(PREC_LOWEST, state)
	if value == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	state.nextToken()

	// "as"
	if !state.expectTokenIs(token.AS) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	asPos := state.curTokenSourceRange()
	state.nextToken()

	// "identifier"
	name1 := p.parseIdent(state)
	if name1 == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	state.nextToken()

	var name2 *ast.Ident
	var commaPos *position.SourceRange
	if state.curTokenIs(token.COMMA) {
		loc := state.curTokenSourceRange()
		commaPos = &loc

		state.nextToken()
		name2 = p.parseIdent(state)
		if name2 == nil {
			return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
		}
		state.nextToken()
	}

	// "{"
	if !state.expectTokenIs(token.LBRACE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	quant := p.parseExpression(PREC_LOWEST, state)
	if quant == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	state.nextToken()

	// It's possible there's an ";" or newline before "}" so consume that.
	if state.curTokenIs(token.SEMICOLON) {
		state.nextToken()
	}

	// "}"
	if !state.expectTokenIs(token.RBRACE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rpos := state.curTokenSourceRange()

	return &ast.QuantExpression{
		NodePos:    state.newSourceRangePosPos(opPos.Start, rpos.End),
		Op:         op,
		OpPos:      opPos,
		Value:      value,
		AsPos:      asPos,
		Name1:      name1,
		CommaPos:   commaPos,
		Name2:      name2,
		LeftBrace:  lpos,
		Quantifier: quant,
		RightBrace: rpos,
	}
}

// RuleExpr = "rule" [ "when" BoolExpr ] "{" Expr "}" .
func (p *parser) parseRuleExpression(state *parserState) ast.Expression {
	state.traceMethodStart("parseRuleLiteral")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "rule"
	if !state.expectTokenIs(token.RULE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rulePos := state.curTokenSourceRange()

	// Documentation
	doc := state.getTrailingComments(rulePos.Start.Byte)

	state.nextToken()

	// When clause
	var whenTokenPos *position.SourceRange
	var whenClause ast.Expression
	if state.curTokenIs(token.WHEN) {
		state.traceMessage("WHEN expressions")
		loc := state.curTokenSourceRange()
		whenTokenPos = &loc
		state.nextToken()
		whenClause = p.parseExpression(PREC_LOWEST, state)
		state.nextToken()
		state.consumeEndOfStatements()
	}

	// "{"
	if !state.expectTokenIs(token.LBRACE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	// Rule value
	value := p.parseExpression(PREC_LOWEST, state)
	state.nextToken()
	state.consumeEndOfStatements()

	// "}"
	if !state.expectTokenIs(token.RBRACE) {
		return &ast.BadExpression{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}
	rpos := state.curTokenSourceRange()

	return &ast.RuleExpression{
		NodePos:       state.newSourceRangeIntInt(startPos, state.curTokenEndPos),
		RuleTokenPos:  rulePos,
		Doc:           doc,
		When:          whenClause,
		WhenTokenPos:  whenTokenPos,
		LeftBracePos:  lpos,
		Value:         value,
		RightBracePos: rpos,
	}
}

func (p *parser) parseSelectorExpression(value ast.Expression, state *parserState) ast.Expression {
	state.traceMethodStart("parseSelectorExpression")
	defer state.traceMethodEnd()

	startPos := value.Position().Start

	// "."
	if !state.expectTokenIs(token.PERIOD) {
		return &ast.BadExpression{NodePos: state.newSourceRangePosInt(startPos, state.curTokenEndPos)}
	}
	state.nextToken()

	sel := p.parseSelectorIdent(state)
	if sel == nil {
		return &ast.BadExpression{NodePos: state.newSourceRangePosInt(startPos, state.curTokenEndPos)}
	}

	return &ast.SelectorExpression{
		NodePos:  state.newSourceRangePosPos(startPos, sel.NodePos.End),
		Value:    value,
		Selector: sel,
	}
}

func (p *parser) parseStringLiteral(state *parserState) ast.Expression {
	state.traceMethodStart("parseStringLiteral")
	defer state.traceMethodEnd()

	if !state.expectTokenIs(token.STRING) {
		return nil
	}

	return &ast.BasicLit{
		NodePos: state.curTokenSourceRange(),
		Kind:    state.curToken.Type,
		Value:   state.curToken.Literal,
	}
}

func (p *parser) parseTrailingUnaryExpression(value ast.Expression, state *parserState) ast.Expression {
	// Trailing unary functions e.g. <value> `is empty`
	return &ast.UnaryExpression{
		NodePos:   state.newSourceRangePosInt(value.Position().Start, state.curTokenEndPos),
		Op:        state.curToken.Type,
		OpPos:     state.curTokenSourceRange(),
		RightExpr: value,
	}
}

func (p *parser) parseUnaryExpression(state *parserState) ast.Expression {
	startPos := state.curTokenStartPos
	op := state.curToken.Type
	opPos := state.curTokenSourceRange()

	state.nextToken()

	rhs := p.parseExpression(PREC_UNARY, state)

	return &ast.UnaryExpression{
		NodePos:   state.newSourceRangeIntPos(startPos, rhs.Position().End),
		Op:        op,
		OpPos:     opPos,
		RightExpr: rhs,
	}
}
