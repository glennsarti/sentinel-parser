package token

import "github.com/glennsarti/sentinel-parser/features"

type TokenType string

const (
	ILLEGAL = "ILLEGAL"

	EOF     = "EOF"
	COMMENT = "COMMENT"

	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"

	ADD = "+"
	SUB = "-"
	MUL = "*"
	QUO = "/"
	REM = "%"

	ADD_ASSIGN = "+="
	SUB_ASSIGN = "-="
	MUL_ASSIGN = "*="
	QUO_ASSIGN = "/="
	REM_ASSIGN = "%="

	EQL = "=="
	LSS = "<"
	GTR = ">"
	NEQ = "!="
	LEQ = "<="
	GEQ = ">="

	ASSIGN = "="

	NOT    = "!"
	NOTSTR = "not"

	LPAREN = "("
	LBRACK = "["
	LBRACE = "{"

	RPAREN    = ")"
	RBRACK    = "]"
	RBRACE    = "}"
	COMMA     = ","
	PERIOD    = "."
	SEMICOLON = ";"
	COLON     = ":"

	LAND = "and"
	LOR  = "or"
	LXOR = "xor"

	CONTAINS    = "contains"
	NOTCONTAINS = "not contains"
	IN          = "in"
	NOTIN       = "not in"
	MATCHES     = "matches"
	NOTMATCHES  = "not matches"
	IS          = "is"
	ISNOT       = "is not"

	IMPORT  = "import"
	PARAM   = "param"
	AS      = "as"
	DEFAULT = "default"
	WHEN    = "when"

	FUNC     = "func"
	RULE     = "rule"
	RETURN   = "return"
	BREAK    = "break"
	CONTINUE = "continue"

	IF     = "if"
	ELSE   = "else"
	ANY    = "any"
	ALL    = "all"
	FOR    = "for"
	FILTER = "filter"
	MAP    = "map"

	CASE = "case"

	EMPTY      = "empty"
	ISEMPTY    = "is empty"
	ISNOTEMPTY = "is not empty"

	DEFINED      = "defined"
	ISDEFINED    = "is defined"
	ISNOTDEFINED = "is not defined"
)

type Token struct {
	Type    TokenType
	Literal string
}

type versionedTokenType struct {
	TokenType
	MinimumVersion string
}

var allKeywords = map[string]versionedTokenType{
	"not": {TokenType: NOTSTR},
	"and": {TokenType: LAND},
	"or":  {TokenType: LOR},
	"xor": {TokenType: LXOR},

	"contains": {TokenType: CONTAINS},
	"in":       {TokenType: IN},
	"matches":  {TokenType: MATCHES},
	"is":       {TokenType: IS},
	"is not":   {TokenType: ISNOT},

	"import":  {TokenType: IMPORT},
	"param":   {TokenType: PARAM},
	"as":      {TokenType: AS},
	"default": {TokenType: DEFAULT},
	"when":    {TokenType: WHEN},

	"func":     {TokenType: FUNC},
	"rule":     {TokenType: RULE},
	"return":   {TokenType: RETURN},
	"break":    {TokenType: BREAK},
	"continue": {TokenType: CONTINUE},

	"if":     {TokenType: IF},
	"else":   {TokenType: ELSE},
	"any":    {TokenType: ANY},
	"all":    {TokenType: ALL},
	"for":    {TokenType: FOR},
	"filter": {TokenType: FILTER},
	"map":    {TokenType: MAP},

	"case": {TokenType: CASE},

	"empty":        {TokenType: EMPTY},
	"is empty":     {TokenType: ISEMPTY},
	"is not empty": {TokenType: ISNOTEMPTY},

	"defined":        {TokenType: DEFINED, MinimumVersion: features.IsDefinedFunctionsMinimumVerions},
	"is defined":     {TokenType: ISDEFINED, MinimumVersion: features.IsDefinedFunctionsMinimumVerions},
	"is not defined": {TokenType: ISNOTDEFINED, MinimumVersion: features.IsDefinedFunctionsMinimumVerions},
}

func LookupIdent(sentinelVersion, ident string) TokenType {
	if tok, ok := allKeywords[ident]; ok {
		if tok.MinimumVersion == "" || features.SupportedVersion(sentinelVersion, tok.MinimumVersion) {
			return tok.TokenType
		}
	}

	return IDENT
}

func SingleWordKeyword(sentinelVersion string, tt TokenType) bool {
	switch tt {
	case ALL,
		ANY,
		AS,
		BREAK,
		CASE,
		CONTAINS,
		CONTINUE,
		DEFAULT,
		ELSE,
		EMPTY,
		FILTER,
		FOR,
		FUNC,
		IF,
		IMPORT,
		IN,
		IS,
		LAND,
		LOR,
		LXOR,
		MAP,
		MATCHES,
		NOTSTR,
		PARAM,
		RETURN,
		RULE,
		WHEN:
		return true
	case DEFINED:
		return features.SupportedVersion(sentinelVersion, features.IsDefinedFunctionsMinimumVerions)
	}
	return false
}
