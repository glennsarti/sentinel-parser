package sentinel

import (
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

type Lexer interface {
	SentinelVersion() string
	NextToken() (int, int, token.Token)
	Locator() token.Locator
}
