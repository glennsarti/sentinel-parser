package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

type fileValidator struct {
	diags diagnostics.Diagnostics
	file  ast.File
}

func ValidateFile(file *ast.File, sentinelVersion string) diagnostics.Diagnostics {
	if file == nil {
		return nil
	}
	v := fileValidator{
		diags: make(diagnostics.Diagnostics, 0),
		file:  *file,
	}

	v.multiImportSchema()
	v.sentinelBlockSupport(sentinelVersion)
	v.paramsInPolicySupport(sentinelVersion)
	v.optionalEnforcementLevelSupport(sentinelVersion)

	return v.diags
}

func (fv *fileValidator) multiImportSchema() {
	if len(fv.file.Imports) <= 1 {
		return
	}
	hasV1 := false
	hasV2 := false

	for _, i := range fv.file.Imports {
		if hasV1 && hasV2 {
			break
		}
		switch i.Schema() {
		case ast.V1ImportSchema:
			hasV1 = true
		case ast.V2ImportSchema:
			hasV2 = true
		}
	}

	// Do we have multiple versions?
	if !hasV1 || !hasV2 {
		return
	}

	for _, i := range fv.file.Imports {
		if i.Schema() == ast.V1ImportSchema {
			continue
		}

		fv.diags = fv.diags.Add(&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Mixed import schema types",
			Detail:   "Cannot use different import schemas in the same configuration",
			Range:    i.Range(),
		})
	}
}

func (fv *fileValidator) optionalEnforcementLevelSupport(sentinelVersion string) {
	if features.SupportedVersion(sentinelVersion, features.OptionalEnforcementLevelMinimumVersion) {
		return
	}

	for _, pol := range fv.file.Policies {
		if pol.EnforcementLevel == "" {
			fv.diags = fv.diags.Add(&diagnostics.Diagnostic{
				Severity: diagnostics.Error,
				Summary:  "Expected attribute",
				Detail:   fmt.Sprintf("enforcement_level attribute in a policy block is not optional in Sentinel %s (Requires %s+).", sentinelVersion, features.OptionalEnforcementLevelMinimumVersion),
				Range:    pol.PolicyRange,
			})
		}
	}
}

func (fv *fileValidator) paramsInPolicySupport(sentinelVersion string) {
	if features.SupportedVersion(sentinelVersion, features.ParamsInPolicyMinimumVersion) {
		return
	}

	for _, pol := range fv.file.Policies {
		if len(pol.Params) > 0 {
			fv.diags = fv.diags.Add(&diagnostics.Diagnostic{
				Severity: diagnostics.Error,
				Summary:  "Unexpected attribute",
				Detail:   fmt.Sprintf("params attribute in a policy block is not valid in Sentinel %s (Requires %s+).", sentinelVersion, features.ParamsInPolicyMinimumVersion),
				Range:    pol.ParamsRange,
			})
		}
	}
}

func (fv *fileValidator) sentinelBlockSupport(sentinelVersion string) {
	if fv.file.SentinelOptions == nil || features.SupportedVersion(sentinelVersion, features.SentinelBlockMinimumVersion) {
		return
	}

	fv.diags = fv.diags.Add(&diagnostics.Diagnostic{
		Severity: diagnostics.Error,
		Summary:  "Unexpected block",
		Detail:   fmt.Sprintf("sentinel block is not valid in Sentinel %s (Requires %s+).", sentinelVersion, features.SentinelBlockMinimumVersion),
		Range:    fv.file.SentinelOptions.SentinelOptionsRange,
	})
}
