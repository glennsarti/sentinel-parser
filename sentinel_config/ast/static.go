package ast

import (
	"github.com/glennsarti/sentinel-parser/position"
)

type V2ImportKind string

const (
	V2ImportKindModule = "module"
	V2ImportKindPlugin = "plugin"
	V2ImportKindStatic = "static"
)

type ImportSchema string

const (
	V1ImportSchema ImportSchema = "v1"
	V2ImportSchema ImportSchema = "v2"
)

// The base Node interface
type Node interface {
	Range() *position.SourceRange
}

// All HCL nodes implement this
type HCLNode interface {
	Node

	BlockName() string
	BlockNameRange() *position.SourceRange
	BlockType() string
}

// All Imports implement this
type Import interface {
	HCLNode

	Schema() ImportSchema
}

// type SourcePos struct {
// 	Line   int `json:"line"`   // Base 0 Line index
// 	Column int `json:"column"` // Base 0 Column index
// 	Byte   int `json:"byte"`   // Byte offset in the source file
// }

// type SourceRange struct {
// 	Filename string    `json:"filename"`
// 	Start    SourcePos `json:"start"`
// 	End      SourcePos `json:"end"`
// }

func NewFile() *File {
	return &File{
		SentinelOptions: nil,
		Globals:         make(map[string]*Global, 0),
		Imports:         make(map[string]Import, 0),
		Mocks:           make(map[string]*Mock, 0),
		Params:          make(map[string]*Parameter, 0),
		Policies:        make(map[string]*Policy, 0),
		Test:            nil,
	}
}
