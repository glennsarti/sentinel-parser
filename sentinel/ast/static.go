package ast

// TODO: This should be autogenerated!!!!
func NewFile() *File {
	return &File{
		Imports:    make([]*ImportDecl, 0),
		Params:     make([]*ParamDecl, 0),
		Statements: make([]Statement, 0),
	}
}
