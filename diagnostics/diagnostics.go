package diagnostics

import (
	"fmt"
)

var _ error = Diagnostics{}

type Diagnostics []*Diagnostic

func EmptyDiags() Diagnostics {
	return make(Diagnostics, 0)
}

func (d Diagnostics) Error() string {
	count := len(d)
	switch count {
	case 0:
		return "no errors"
	case 1:
		return d[0].Error()
	case 2:
		return fmt.Sprintf("%s, and one other diagnostic", d[0].Error())
	default:
		return fmt.Sprintf("%s, and %d other diagnostics", d[0].Error(), count-1)
	}
}

func (d Diagnostics) HasErrors() bool {
	for _, diag := range d {
		if diag != nil && diag.Severity == Error {
			return true
		}
	}
	return false
}

func (d Diagnostics) Add(diag *Diagnostic) Diagnostics {
	if diag != nil {
		return append(d, diag)
	} else {
		return d
	}
}

func (d Diagnostics) Concat(diags Diagnostics) Diagnostics {
	return append(d, diags...)
}

func (d Diagnostics) Errors() []*Diagnostic {
	result := make([]*Diagnostic, 0)
	for _, diag := range d {
		if diag != nil && diag.Severity == Error {
			result = append(result, diag)
		}
	}
	return result
}
