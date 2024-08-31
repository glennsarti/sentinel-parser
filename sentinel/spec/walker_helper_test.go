package spec

import "github.com/glennsarti/sentinel-parser/sentinel/ast"

type SpecTestWalker struct {
	NodesVisited int
}

func (w *SpecTestWalker) Visit(n ast.Node) ast.VisitFunc {
	if n != nil {
		// Only increment for actual nodes
		w.NodesVisited += 1
	}
	return w.Visit
}
