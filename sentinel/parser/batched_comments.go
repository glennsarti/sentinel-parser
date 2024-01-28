package parser

import (
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
)

// Helper to manage list of comments while parsing
type batchedComments struct {
	Comments []*ast.Comment
}

func (bc *batchedComments) LeadingComments(fromLine int) *ast.Comments {
	if len(bc.Comments) == 0 {
		return nil
	}

	var cg *ast.Comments = nil

	for _, comment := range bc.Comments {
		thisLine := comment.NodePos.End.Line
		if fromLine == thisLine || fromLine == token.NoPos {
			if cg == nil {
				cg = &ast.Comments{
					List:    []*ast.Comment{comment},
					NodePos: comment.NodePos.Clone(),
				}
			} else {
				cg.List = append(cg.List, comment)
				cg.NodePos.End = comment.NodePos.End
			}
			fromLine = thisLine + 1
		} else {
			return cg
		}
	}
	return cg
}

func (bc *batchedComments) TrailingComments(linenum int) *ast.Comments {
	if len(bc.Comments) == 0 {
		return nil
	}

	var cg *ast.Comments = nil
	for idx := len(bc.Comments) - 1; idx >= 0; idx-- {
		comment := bc.Comments[idx]
		if linenum == comment.NodePos.End.Line {
			if cg == nil {
				cg = &ast.Comments{
					List:    []*ast.Comment{comment},
					NodePos: comment.NodePos.Clone(),
				}
			} else {
				// Prepend, not append
				cg.List = append([]*ast.Comment{comment}, cg.List...)
				cg.NodePos.Start = comment.NodePos.Start
			}
		} else {
			return cg
		}
		linenum = comment.NodePos.Start.Line - 1
	}
	return cg
}
