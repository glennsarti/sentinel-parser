package parser

import (
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

func (p *parser) parseStatement(state *parserState) ast.Statement {
	state.traceMethodStart("ParseStatement")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	switch state.curToken.Type {
	case token.LBRACE:
		if s := p.parseBlockStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.RETURN:
		if s := p.parseReturnStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.BREAK:
		if s := p.parseBreakStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.CONTINUE:
		if s := p.parseContinueStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.IF:
		if s := p.parseIfStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.FOR:
		if s := p.parseForStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.FUNC:
		if p.features.FuncDeclarations {
			if s := p.parseFuncDecl(state); s != nil {
				return s
			} else {
				return p.nilBadStatement(startPos, state)
			}
		}
	case token.CASE:
		if s := p.parseCaseStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	case token.SEMICOLON:
		if s := p.parseEmptyStatement(state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	}

	// If it's not an explicit statement then it's an Expression
	expr := p.parseExpression(PREC_LOWEST, state)
	if expr == nil {
		return &ast.BadStatement{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	// Is this an assignment?
	if p.isAssignOpToken(state.peekToken.Type) {
		state.nextToken()
		if s := p.parseAssignStatement(expr, state); s != nil {
			return s
		} else {
			return p.nilBadStatement(startPos, state)
		}
	}

	if !state.expectPeekEndOfStmt() {
		return &ast.BadStatement{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
	}

	// Must be an expression statement then
	return &ast.ExpressionStatement{
		NodePos: expr.Position(),
		Expr:    expr,
	}
}

func (p *parser) nilBadStatement(startPos int, state *parserState) ast.Statement {
	return &ast.BadStatement{NodePos: state.newSourceRangeIntInt(startPos, state.curTokenEndPos)}
}

// Assignment = AssignExpr assign_op Expression .
// AssignExpr = identifier | IndexExpr .
// assign_op  = [ add_op | mul_op ] "=" .
func (p *parser) parseAssignStatement(lhs ast.Expression, state *parserState) ast.Statement {
	state.traceMethodStart("ParseAssignStatement")
	defer state.traceMethodEnd()

	if !p.isAssignOpToken(state.curToken.Type) {
		return &ast.BadStatement{NodePos: state.newSourceRangePosInt(lhs.Position().Start, state.curTokenEndPos)}
	}
	assignOp := state.curToken.Type
	assignPos := state.curTokenSourceRange()
	state.nextToken()

	rhs := p.parseExpression(PREC_LOWEST, state)
	if rhs == nil {
		return &ast.BadStatement{NodePos: state.newSourceRangePosInt(lhs.Position().Start, state.curTokenEndPos)}
	}

	if !state.expectPeekEndOfStmt() {
		return &ast.BadStatement{NodePos: state.newSourceRangePosInt(lhs.Position().Start, state.curTokenEndPos)}
	}

	return &ast.AssignStatement{
		NodePos:     state.newSourceRangePosPos(lhs.Position().Start, rhs.Position().End),
		LeftExpr:    lhs,
		AssignOp:    assignOp,
		AssignOpPos: assignPos,
		RightExpr:   rhs,
	}
}

func (p *parser) parseBlockStatement(state *parserState) *ast.BlockStatement {
	state.traceMethodStart("parseBlockStatement")
	defer state.traceMethodEnd()

	// "{"
	if !state.expectTokenIs(token.LBRACE) {
		return nil
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	list := make([]ast.Statement, 0)
	for !state.curTokenIs(token.RBRACE) && !state.curTokenIs(token.EOF) {
		Statement := p.parseStatement(state)
		if Statement == nil {
			return nil
		}

		list = append(list, Statement)

		if !state.curTokenIs(token.SEMICOLON) && state.peekTokenIs(token.SEMICOLON) {
			state.nextToken()
			state.nextToken()
		} else {
			state.nextToken()
		}
	}

	// "}"
	if !state.expectTokenIs(token.RBRACE) {
		return nil
	}
	rpos := state.curTokenSourceRange()

	return &ast.BlockStatement{
		NodePos:    state.newSourceRangePosPos(lpos.Start, rpos.End),
		LeftBrace:  lpos,
		Statments:  list,
		RightBrace: rpos,
	}
}

// BreakStatement = "break" .
func (p *parser) parseBreakStatement(state *parserState) *ast.BranchStatement {
	state.traceMethodStart("parseBreakStatement")
	defer state.traceMethodEnd()

	return p.parseBranchStatement(token.BREAK, state)
}

// CaseStatement = "case" [ Expression ] "{" { CaseWhenClause } "}" .
func (p *parser) parseCaseStatement(state *parserState) *ast.CaseStatement {
	state.traceMethodStart("parseCaseStatement")
	defer state.traceMethodEnd()

	// "case"
	if !state.expectTokenIs(token.CASE) {
		return nil
	}
	casePos := state.curTokenSourceRange()
	state.nextToken()

	var caseValue ast.Expression = nil
	// "{"
	if !state.curTokenIs(token.LBRACE) {
		// <expr> "{"
		caseValue = p.parseExpression(PREC_LOWEST, state)
		state.nextToken()
		if !state.expectTokenIs(token.LBRACE) {
			return nil
		}
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	clauses := make([]ast.Statement, 0)
	for !state.curTokenIs(token.RBRACE) && !state.curTokenIs(token.EOF) {
		if Statement := p.parseCaseWhenClause(state); Statement == nil {
			return nil
		} else {
			clauses = append(clauses, Statement)
		}

		if !state.curTokenIs(token.SEMICOLON) && state.peekTokenIs(token.SEMICOLON) {
			state.nextToken()
			state.nextToken()
		} else {
			state.nextToken()
		}
	}

	// "}"
	if !state.expectTokenIs(token.RBRACE) {
		return nil
	}
	rpos := state.curTokenSourceRange()

	state.consumePeekSemicolon()

	return &ast.CaseStatement{
		NodePos: state.newSourceRangePosPos(casePos.Start, rpos.End),
		CasePos: casePos,
		Value:   caseValue,
		Clauses: &ast.BlockStatement{
			NodePos:    state.newSourceRangePosPos(lpos.Start, rpos.End),
			LeftBrace:  lpos,
			Statments:  clauses,
			RightBrace: rpos,
		},
	}
}

// CaseWhenClause = CaseWhenCase ":" StatementList .
// CaseWhenCase = "when" Expression { "," Expression } | "else" .
func (p *parser) parseCaseWhenClause(state *parserState) *ast.CaseWhenClause {
	state.traceMethodStart("parseCaseWhenClause")
	defer state.traceMethodEnd()

	// "when" or "else"
	startTok := state.curToken.Type
	tokenPos := state.curTokenSourceRange()
	conds := make([]ast.Expression, 0)
	switch state.curToken.Type {
	case token.WHEN:
		conds = p.parseExpressionList(token.COLON, state)
	case token.ELSE:
		state.nextToken()
	default:
		state.expectTokenIs(token.WHEN)
		return nil
	}

	if !state.expectTokenIs(token.COLON) {
		return nil
	}

	colonPos := state.curTokenSourceRange()
	Statements := make([]ast.Statement, 0)
	endPos := colonPos.Start.Clone()

	for !state.peekTokenIs(token.WHEN) && !state.peekTokenIs(token.ELSE) && !state.peekTokenIs(token.RBRACE) && !state.peekTokenIs(token.EOF) {
		state.nextToken()
		if Statement := p.parseStatement(state); Statement == nil {
			return nil
		} else {
			Statements = append(Statements, Statement)
			endPos = Statement.Position().End
		}
	}

	return &ast.CaseWhenClause{
		NodePos:    state.newSourceRangePosPos(tokenPos.Start, endPos),
		TokenKind:  startTok,
		TokenPos:   tokenPos,
		Conditions: conds,
		ColonPos:   colonPos,
		Statements: Statements,
	}
}

// ContinueStatement = "continue" .
func (p *parser) parseContinueStatement(state *parserState) *ast.BranchStatement {
	state.traceMethodStart("parseContinueStatement")
	defer state.traceMethodEnd()

	return p.parseBranchStatement(token.CONTINUE, state)
}

func (p *parser) parseEmptyStatement(state *parserState) *ast.EmptyStatement {
	state.traceMethodStart("parseEmptyStatement")
	defer state.traceMethodEnd()

	if !state.expectTokenIs(token.SEMICOLON) {
		return nil
	}

	return &ast.EmptyStatement{
		NodePos: state.curTokenSourceRange(),
		Implied: state.curToken.Literal != ";",
	}
}

// FunctionDecl    = "func" identifier Function .
// Function       = Parameters FunctionBody .
// Parameters     = "(" [ IdentifierList [ "," ] ] ")" .
// FunctionBody   = Block .
// IdentifierList = identifier { "," identifier } .
func (p *parser) parseFuncDecl(state *parserState) *ast.FuncDecl {
	state.traceMethodStart("parseFuncDecl")
	defer state.traceMethodEnd()

	// "func"
	if !state.expectTokenIs(token.FUNC) {
		return nil
	}
	fpos := state.curTokenSourceRange()

	// Documentation
	doc := state.getTrailingComments(fpos.Start.Byte)

	state.nextToken()

	name := p.parseIdent(state)
	if name == nil {
		return nil
	}
	state.nextToken()

	fields := p.parseFieldList(state)
	if fields == nil {
		return nil
	}
	state.nextToken()

	body := p.parseBlockStatement(state)
	if body == nil {
		return nil
	}

	state.consumePeekSemicolon()

	return &ast.FuncDecl{
		NodePos: state.newSourceRangePosPos(fpos.Start, body.Position().End),
		Doc:     doc,
		FuncPos: fpos,
		Name:    name,
		Params:  fields,
		Body:    body,
	}
}

// IfStatement = "if" BoolExpr Block [ "else" ( IfStatement | Block ) ] .
func (p *parser) parseIfStatement(state *parserState) *ast.IfStatement {
	state.traceMethodStart("parseIfStatement")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "if"
	if !state.expectTokenIs(token.IF) {
		return nil
	}
	ifPos := state.curTokenSourceRange()
	state.nextToken()

	cond := p.parseExpression(PREC_LOWEST, state)
	if cond == nil {
		return nil
	}
	state.nextToken()

	trueBlock := p.parseBlockStatement(state)
	if trueBlock == nil {
		return nil
	}
	endPos := trueBlock.Position().End

	var falseBlock ast.Statement = nil
	elsePos := state.newInvalidSourceRange()

	switch state.peekToken.Type {
	case token.SEMICOLON:
		// If it finished when a semicolon (`{...};`) then there can't be an else
		state.nextToken()
		return &ast.IfStatement{
			NodePos:    state.newSourceRangeIntPos(startPos, endPos),
			IfPos:      ifPos,
			Condition:  cond,
			TrueBlock:  trueBlock,
			ElsePos:    elsePos,
			FalseBlock: falseBlock,
		}
	case token.ELSE:
		// {...} else {...}
		state.traceMessage("Found 'else' keyword")
		state.nextToken()
		elsePos = state.curTokenSourceRange()
		state.nextToken()

		switch state.curToken.Type {
		case token.IF:
			if b := p.parseIfStatement(state); b == nil {
				return nil
			} else {
				falseBlock = b
			}
		case token.LBRACE:
			if b := p.parseBlockStatement(state); b == nil {
				return nil
			} else {
				falseBlock = b
			}
			state.consumePeekSemicolon()
		default:
			state.addExpectedStringParserError("if or block")
			falseBlock = &ast.BadStatement{NodePos: state.curTokenSourceRange()}
		}
		endPos = falseBlock.Position().End

		return &ast.IfStatement{
			NodePos:    state.newSourceRangeIntPos(startPos, endPos),
			IfPos:      ifPos,
			Condition:  cond,
			TrueBlock:  trueBlock,
			ElsePos:    elsePos,
			FalseBlock: falseBlock,
		}
	default:
		state.nextToken()
		state.expectTokenIs(token.ELSE)
		return nil
	}
}

// ForStatement = "for" Expressions "as" [ identifier "," ] identifier Block .
func (p *parser) parseForStatement(state *parserState) *ast.ForStatement {
	state.traceMethodStart("parseForStatement")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "for"
	if !state.expectTokenIs(token.FOR) {
		return nil
	}
	forPos := state.curTokenSourceRange()
	state.nextToken()

	iterable := p.parseExpression(PREC_LOWEST, state)
	if iterable == nil {
		return nil
	}
	state.nextToken()

	// "as"
	if !state.expectTokenIs(token.AS) {
		return nil
	}
	asPos := state.curTokenSourceRange()
	state.nextToken()

	iter1 := p.parseIdent(state)
	if iter1 == nil {
		return nil
	}
	state.nextToken()

	var iter2 *ast.Ident = nil
	if state.curTokenIs(token.COMMA) {
		state.nextToken()

		iter2 = p.parseIdent(state)
		if iter2 == nil {
			return nil
		}
		state.nextToken()
	}

	block := p.parseBlockStatement(state)
	if block == nil {
		return nil
	}
	state.expectPeekEndOfStmt()

	return &ast.ForStatement{
		NodePos:   state.newSourceRangeIntPos(startPos, block.NodePos.End),
		ForPos:    forPos,
		Iterable:  iterable,
		AsPos:     asPos,
		Iterator1: iter1,
		Iterator2: iter2,
		Block:     block,
	}
}

// ReturnStatement = "return" Expression .
func (p *parser) parseReturnStatement(state *parserState) *ast.ReturnStatement {
	state.traceMethodStart("parseReturnStatement")
	defer state.traceMethodEnd()

	if !state.expectTokenIs(token.RETURN) {
		return nil
	}
	rpos := state.curTokenSourceRange()
	state.nextToken()

	result := p.parseExpression(PREC_LOWEST, state)

	if !state.expectPeekEndOfStmt() {
		return nil
	}

	return &ast.ReturnStatement{
		NodePos:   state.newSourceRangePosPos(rpos.Start, result.Position().End),
		ReturnPos: rpos,
		Result:    result,
	}
}

func (p *parser) parseFieldList(state *parserState) *ast.FieldList {
	state.traceMethodStart("parseFieldList")
	defer state.traceMethodEnd()

	// "("
	if !state.expectTokenIs(token.LPAREN) {
		return nil
	}
	lpos := state.curTokenSourceRange()
	state.nextToken()

	list := make([]*ast.Ident, 0)
	if state.curTokenIs(token.RPAREN) {
		return &ast.FieldList{
			NodePos:    state.newSourceRangePosInt(lpos.Start, state.curTokenStartPos),
			LeftParen:  lpos,
			Fields:     list,
			RightParen: state.curTokenSourceRange(),
		}
	}

	// Get the 1st item
	if ident := p.parseIdent(state); ident == nil {
		return nil
	} else {
		list = append(list, ident)
	}

	// Field lists can terminate with
	// ( <Ident> )
	// OR
	// (
	//   ...
	//   <Ident>,  <--- Mandatory trailing comma on multiline lists
	// )
	for state.peekTokenIs(token.COMMA) {
		state.nextToken()

		if state.peekTokenIs(token.RPAREN) {
			// Terminate if this is a trailing comma; "," ")"
			break
		}
		state.nextToken()

		if ident := p.parseIdent(state); ident == nil {
			return nil
		} else {
			list = append(list, ident)
		}
	}
	state.nextToken()

	// ")"
	if !state.expectTokenIs(token.RPAREN) {
		return nil
	}
	rpos := state.curTokenSourceRange()

	return &ast.FieldList{
		NodePos:    state.newSourceRangePosPos(lpos.Start, rpos.End),
		LeftParen:  lpos,
		Fields:     list,
		RightParen: rpos,
	}
}

func (p *parser) parseBranchStatement(tt token.TokenType, state *parserState) *ast.BranchStatement {
	state.traceMethodStart("parseBranchStatement")
	defer state.traceMethodEnd()

	if state.expectTokenIs(tt) {
		return &ast.BranchStatement{
			NodePos: state.curTokenSourceRange(),
			Kind:    tt,
		}
	} else {
		return nil
	}
}

func (p *parser) isAssignOpToken(tt token.TokenType) bool {
	return tt == token.ASSIGN ||
		tt == token.ADD_ASSIGN ||
		tt == token.SUB_ASSIGN ||
		tt == token.MUL_ASSIGN ||
		tt == token.QUO_ASSIGN ||
		tt == token.REM_ASSIGN
}
