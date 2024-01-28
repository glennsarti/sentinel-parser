package diagnostics

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/position"
)

var _ error = Diagnostic{}

type Diagnostic struct {
	Severity SeverityLevel

	Summary string
	Detail  string
	Range   *position.SourceRange

	Related *[]RelatedDiagnostic
}

type RelatedDiagnostic struct {
	Range   *position.SourceRange
	Summary string
}

func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s: (%s) %s; %s", d.Severity.ToString(), d.Range.ToString(), d.Summary, d.Detail)
}
