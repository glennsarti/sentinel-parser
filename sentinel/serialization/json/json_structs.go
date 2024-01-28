package json

type jsonSourceRange struct {
	End      jsonSourcePos `json:"end"`
	Filename string        `json:"filename"`
	Start    jsonSourcePos `json:"start"`
}

type jsonSourcePos struct {
	Byte   int `json:"byte"`
	Column int `json:"column"`
	Line   int `json:"line"`
}
