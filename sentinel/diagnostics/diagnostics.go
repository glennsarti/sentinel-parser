package diagnostics

// import (
// 	"github.com/glennsarti/sentinel-parser/sentinel/ast"
// )

// type ParserErrors struct {
// 	list []ParserError
// }

// type ParserError struct {
// 	Message  string
// 	Location ast.SourceRange
// }

// func NewParserErrors() *ParserErrors {
// 	return &ParserErrors{
// 		list: make([]ParserError, 0),
// 	}
// }

// func (pe *ParserErrors) Add(message string, loc ast.SourceRange) {
// 	pe.list = append(pe.list, ParserError{
// 		Message:  message,
// 		Location: loc,
// 	})
// }

// func (pe *ParserErrors) Errors() []ParserError {
// 	return pe.list
// }
