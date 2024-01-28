package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
	"github.com/glennsarti/sentinel-parser/sentinel/tracing"
)

type parser struct {
	sentinelVersion string
	features        features.SentinelFeatures

	unaryParseFuncs  map[token.TokenType]unaryParseFunc
	binaryParseFuncs map[token.TokenType]binaryParseFunc
	suffixParseFuncs map[token.TokenType]suffixParseFunc
}

func newParser(
	sentinelVersion string,
) (*parser, error) {
	ok, ver := features.ValidateSentinelVersion(sentinelVersion)
	if !ok {
		return nil, fmt.Errorf("%s is an invalid sentinel version", sentinelVersion)
	}

	p := &parser{
		sentinelVersion: ver,
		features:        features.SupportedFeatures(ver),

		unaryParseFuncs:  make(map[token.TokenType]unaryParseFunc),
		binaryParseFuncs: make(map[token.TokenType]binaryParseFunc),
		suffixParseFuncs: make(map[token.TokenType]suffixParseFunc),
	}

	// Unary Functions
	p.registerUnaryFunc(token.INT, p.parseIntLiteral)
	p.registerUnaryFunc(token.FLOAT, p.parseFloatLiteral)
	p.registerUnaryFunc(token.STRING, p.parseStringLiteral)
	p.registerUnaryFunc(token.IDENT, p.parseIdentExpression)
	p.registerUnaryFunc(token.LBRACK, p.parseListLit)
	p.registerUnaryFunc(token.LBRACE, p.parseMapLiteral)
	p.registerUnaryFunc(token.LPAREN, p.parseGroupExpression)
	p.registerUnaryFunc(token.FUNC, p.parseFuncLiteral)
	p.registerUnaryFunc(token.RULE, p.parseRuleExpression)
	// unary_op   = "+" | "-" | "!" | "not" | empty_op | defined_op .
	p.registerUnaryFunc(token.SUB, p.parseUnaryExpression)
	p.registerUnaryFunc(token.NOT, p.parseUnaryExpression)
	p.registerUnaryFunc(token.NOTSTR, p.parseUnaryExpression)
	// Quant Expr
	p.registerUnaryFunc(token.ANY, p.parseQuantExpression)
	p.registerUnaryFunc(token.ALL, p.parseQuantExpression)
	p.registerUnaryFunc(token.MAP, p.parseQuantExpression)
	p.registerUnaryFunc(token.FILTER, p.parseQuantExpression)

	// Binary Functions
	p.registerBinaryFunc(token.ADD, p.parseBinaryExpr)
	p.registerBinaryFunc(token.CONTAINS, p.parseBinaryExpr)
	p.registerBinaryFunc(token.ELSE, p.parseBinaryExpr)
	p.registerBinaryFunc(token.EQL, p.parseBinaryExpr)
	p.registerBinaryFunc(token.GEQ, p.parseBinaryExpr)
	p.registerBinaryFunc(token.GTR, p.parseBinaryExpr)
	p.registerBinaryFunc(token.IN, p.parseBinaryExpr)
	p.registerBinaryFunc(token.IS, p.parseBinaryExpr)
	p.registerBinaryFunc(token.ISNOT, p.parseBinaryExpr)
	p.registerBinaryFunc(token.LAND, p.parseBinaryExpr)
	p.registerBinaryFunc(token.LEQ, p.parseBinaryExpr)
	p.registerBinaryFunc(token.LOR, p.parseBinaryExpr)
	p.registerBinaryFunc(token.LSS, p.parseBinaryExpr)
	p.registerBinaryFunc(token.LXOR, p.parseBinaryExpr)
	p.registerBinaryFunc(token.MATCHES, p.parseBinaryExpr)
	p.registerBinaryFunc(token.MUL, p.parseBinaryExpr)
	p.registerBinaryFunc(token.NEQ, p.parseBinaryExpr)
	p.registerBinaryFunc(token.NOTCONTAINS, p.parseBinaryExpr)
	p.registerBinaryFunc(token.NOTIN, p.parseBinaryExpr)
	p.registerBinaryFunc(token.NOTMATCHES, p.parseBinaryExpr)
	p.registerBinaryFunc(token.NOTSTR, p.parseBinaryExpr)
	p.registerBinaryFunc(token.QUO, p.parseBinaryExpr)
	p.registerBinaryFunc(token.REM, p.parseBinaryExpr)
	p.registerBinaryFunc(token.SUB, p.parseBinaryExpr)

	// Suffix Functions
	// <expr>[....  Could be a slice or an index
	p.registerSuffixFunc(token.LBRACK, p.parseIndexOrSliceExpression)
	// <expr>(.... could be a function call
	p.registerSuffixFunc(token.LPAREN, p.parseCallExpression)
	// <expr>.<selector>
	p.registerSuffixFunc(token.PERIOD, p.parseSelectorExpression)
	// Trailing unary functions e.g. <expr> `is empty`
	p.registerSuffixFunc(token.ISEMPTY, p.parseTrailingUnaryExpression)
	p.registerSuffixFunc(token.ISNOTEMPTY, p.parseTrailingUnaryExpression)
	p.registerSuffixFunc(token.ISDEFINED, p.parseTrailingUnaryExpression)
	p.registerSuffixFunc(token.ISNOTDEFINED, p.parseTrailingUnaryExpression)

	return p, nil
}

// Public Methods
func (p *parser) SentinelVersion() string { return p.sentinelVersion }

func (p *parser) ParseFile(
	lex sentinel.Lexer,
	locator token.Locator,
	tracer tracing.ParsingTracer,
) (*ast.File, diagnostics.Diagnostics, error) {
	if state, err := p.newParserState(lex, locator, tracer); err != nil {
		return nil, state.parserErrors, err
	} else {
		return p.parseFile(state), state.parserErrors, nil
	}
}

func (p *parser) ParseExpression() ast.Expression {
	// TODO: Implement This
	return nil // p.parseExpression(PREC_LOWEST)
}

func (p *parser) ParseStatement() ast.Statement {
	// TODO: Implement This
	return nil // p.parseStatement()
}

func (p *parser) parseFile(state *parserState) *ast.File {
	state.traceMethodStart("ParseFile")
	defer state.traceMethodEnd()

	startPos := state.sourcePosOf(state.curTokenStartPos)
	endPos := startPos.Clone()

	// Documentation
	var doc *ast.Comments = nil
	if state.curBatchedComments != nil {
		// File comments start on the first line comment
		if doc = state.curBatchedComments.LeadingComments(token.NoPos); doc != nil {
			// We need at least one blank line between the file comments
			// and the next token for a valid file comment
			lastCommentLine := doc.NodePos.End.Line
			nextTokenLine := state.lex.Locator().PositionOf(state.curTokenLocation().Start).Line
			if nextTokenLine == lastCommentLine+1 {
				doc = nil
			}
		}
	}

	// Imports at the top of the file
	imports := make([]*ast.ImportDecl, 0)
	for state.curTokenIs(token.IMPORT) {
		state.traceMessage("Parsing imports...")
		if i := p.parseImportDeclaration(state); i != nil {
			imports = append(imports, i)
			endPos = i.NodePos.End.Clone()
		}
		state.nextToken()
	}

	// Params are next
	params := make([]*ast.ParamDecl, 0)
	for state.curTokenIs(token.PARAM) {
		state.traceMessage("Parsing params...")
		if i := p.parseParamDeclaration(state); i != nil {
			params = append(params, i)
			endPos = i.NodePos.End.Clone()
		}
		state.nextToken()
	}

	// The rest are statements
	stmts := make([]ast.Statement, 0)
	for !state.curTokenIs(token.EOF) {
		state.traceMessage("Parsing statement...")
		if s := p.parseStatement(state); s != nil {
			endPos = s.Position().End.Clone()
			stmts = append(stmts, s)
		}
		state.nextToken()
	}

	return &ast.File{
		NodePos:    state.newSourceRangePosPos(startPos, endPos),
		Doc:        doc,
		Imports:    imports,
		Params:     params,
		Statements: stmts,
	}
}

func (p *parser) parseIdent(state *parserState) *ast.Ident {
	state.traceMethodStart("parseIdent")
	defer state.traceMethodEnd()

	if !state.expectTokenIs(token.IDENT) {
		return nil
	}
	result := &ast.Ident{
		NodePos: state.curTokenSourceRange(),
		Name:    state.curToken.Literal,
	}

	return result
}

func (p *parser) parseSelectorIdent(state *parserState) *ast.Ident {
	// The ident on a selector can, unfortunately, be a keyword or an identifier
	// so we need to process it slightly differently e.g.;
	//
	// `foo = bar.func` is valid
	//
	// This should really be `foo = bar["func"]
	state.traceMethodStart("parseSelectorIdent")
	defer state.traceMethodEnd()

	if token.SingleWordKeyword(p.sentinelVersion, state.curToken.Type) {
		return &ast.Ident{
			NodePos: state.curTokenSourceRange(),
			Name:    state.curToken.Literal,
		}
	}

	if !state.expectTokenIs(token.IDENT) {
		return nil
	}
	return &ast.Ident{
		NodePos: state.curTokenSourceRange(),
		Name:    state.curToken.Literal,
	}
}

// ImportDecl = "import" Name ["as" identifier] .
func (p *parser) parseImportDeclaration(state *parserState) *ast.ImportDecl {
	state.traceMethodStart("parseImportDeclaration")
	defer state.traceMethodEnd()

	startPos := state.curTokenStartPos

	// "import"
	if !state.expectTokenIs(token.IMPORT) {
		return nil
	}
	impPos := state.curTokenSourceRange()

	// Documentation
	doc := state.getTrailingComments(impPos.Start.Byte)

	state.nextToken()

	// name
	if !state.expectTokenIs(token.STRING) {
		return nil
	}
	name := state.curToken.Literal
	namePos := state.curTokenSourceRange()
	endPos := namePos.End

	// "as"
	asPos := state.newInvalidSourceRange()
	var alias *ast.Ident
	if state.peekTokenIs(token.AS) {
		state.nextToken()
		asPos = state.curTokenSourceRange()
		state.nextToken()
		alias = p.parseIdent(state)
		if alias == nil {
			return nil
		}
		endPos = alias.NodePos.End
	}

	state.consumePeekEndOfStatements()
	return &ast.ImportDecl{
		NodePos:   state.newSourceRangeIntPos(startPos, endPos),
		Doc:       doc,
		ImportPos: impPos,
		Name:      &ast.BasicLit{NodePos: namePos, Kind: token.STRING, Value: name},
		AsPos:     asPos,
		Alias:     alias,
	}
}

func (p *parser) parseParamDeclaration(state *parserState) (result *ast.ParamDecl) {
	state.traceMethodStart("parseParamDeclaration")
	defer state.traceMethodEnd()

	result = nil
	if !state.expectTokenIs(token.PARAM) {
		return
	}

	paramPos := state.curTokenStartPos

	// Documentation
	doc := state.getTrailingComments(paramPos)

	state.nextToken()
	ident := p.parseIdent(state)
	if ident == nil {
		return nil
	}

	var defExpr ast.Expression = nil
	defaultPos := state.newInvalidSourceRange()
	if state.peekTokenIs(token.DEFAULT) {
		state.nextToken()
		defaultPos = state.curTokenSourceRange()
		state.nextToken()
		defExpr = p.parseExpression(PREC_LOWEST, state)
	}
	endPos := state.curTokenEndPos

	state.consumePeekEndOfStatements()

	result = &ast.ParamDecl{
		NodePos:    state.newSourceRangeIntInt(paramPos, endPos),
		Doc:        doc,
		Name:       ident,
		DefaultPos: defaultPos,
		Default:    defExpr,
	}

	return
}
