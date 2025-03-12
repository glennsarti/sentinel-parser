package parser

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

func OverrideFileWith(this, other *ast.File, sentinelVersion string) diagnostics.Diagnostics {
	return overrideFileWith(this, other, true, sentinelVersion)
}

func AttemptOverrideFileWith(this, other *ast.File, sentinelVersion string) diagnostics.Diagnostics {
	return overrideFileWith(this, other, false, sentinelVersion)
}

func overrideFileWith(this, other *ast.File, overrideValues bool, sentinelVersion string) diagnostics.Diagnostics {

	if features.UnsupportedVersion(sentinelVersion, features.ConfigurationOverrideMinimumVersion) {
		return diagnostics.Diagnostics{&diagnostics.Diagnostic{
			Severity: diagnostics.Error,
			Summary:  "Configuration",
			Detail:   fmt.Sprintf("Configuration overrides are not allowed in Sentinel %s (Requires %s)", sentinelVersion, features.ConfigurationOverrideMinimumVersion),
		}}
	}

	diags := diagnostics.EmptyDiags()

	if this == nil || other == nil {
		return diags
	}

	// Globals
	if len(other.Globals) > 0 {
		d := overrideMap(
			this.Globals,
			other.Globals,
			func(name string) diagnostics.Diagnostics {
				return overrideGlobalWith(this.Globals[name], other.Globals[name], overrideValues)
			},
		)

		diags = append(diags, d...)
	}

	// Imports
	if len(other.Imports) > 0 {
		d := overrideImports(this, other, overrideValues)
		diags = append(diags, d...)
	}

	// Mocks
	if len(other.Mocks) > 0 {
		d := overrideMap(
			this.Mocks,
			other.Mocks,
			func(name string) diagnostics.Diagnostics {
				return overrideMockWith(this.Mocks[name], other.Mocks[name], overrideValues)
			},
		)

		diags = append(diags, d...)
	}

	// Params
	if len(other.Params) > 0 {
		d := overrideMap(
			this.Params,
			other.Params,
			func(name string) diagnostics.Diagnostics {
				return overrideParameterWith(this.Params[name], other.Params[name], overrideValues)
			},
		)

		diags = append(diags, d...)
	}

	// Policies
	if len(other.Policies) > 0 {
		d := overrideMap(
			this.Policies,
			other.Policies,
			func(name string) diagnostics.Diagnostics {
				return overridePolicyWith(this.Policies[name], other.Policies[name], overrideValues)
			},
		)

		diags = append(diags, d...)
	}

	// Sentinel Options
	if other.SentinelOptions != nil {
		d := overrideSentinelOptions(this, other, overrideValues)
		diags = append(diags, d...)
	}

	// Test
	if overrideValues && other.Test != nil {
		this.Test = ast.CloneTest(other.Test)
	}
	return diags
}

// Object overriders
func overrideGlobalWith(this, other *ast.Global, overrideValues bool) diagnostics.Diagnostics {
	if this == nil {
		return diagnostics.EmptyDiags()
	}

	if other != nil {
		if overrideValues && other.ValueRange != nil {
			this.Value = ast.CloneDynamicValue(other.Value)
			this.ValueRange = ast.CloneRange(other.ValueRange)
			this.ValueType = other.ValueType
		}
	}

	return diagnostics.EmptyDiags()
}

func overrideMockWith(this, other *ast.Mock, overrideValues bool) diagnostics.Diagnostics {
	if this == nil {
		return diagnostics.EmptyDiags()
	}

	if overrideValues && other.DataRange != nil {
		this.DataRange = ast.CloneRange(other.DataRange)

		this.Data = ast.CloneParameterMap(other.Data)
	}

	if overrideValues && other.Module != nil {
		this.Module = ast.CloneMockModule(other.Module)
	}

	return diagnostics.EmptyDiags()
}

func overrideParameterWith(this, other *ast.Parameter, overrideValues bool) diagnostics.Diagnostics {
	if this == nil {
		return diagnostics.EmptyDiags()
	}

	if overrideValues && other.ValueRange != nil {
		this.Value = ast.CloneDynamicValue(other.Value)
		this.ValueRange = ast.CloneRange(other.ValueRange)
		this.ValueType = other.ValueType
	}

	return diagnostics.EmptyDiags()
}

func overridePolicyWith(this, other *ast.Policy, overrideValues bool) diagnostics.Diagnostics {
	if this == nil {
		return diagnostics.EmptyDiags()
	}

	if overrideValues && other.SourceRange != nil {
		this.Source = other.Source
		this.SourceRange = ast.CloneRange(other.SourceRange)
	}

	if overrideValues && other.EnforcementLevelRange != nil {
		this.EnforcementLevel = other.EnforcementLevel
		this.EnforcementLevelRange = ast.CloneRange(other.EnforcementLevelRange)
	}

	if overrideValues && other.ParamsRange != nil {
		this.Params = other.Params
		this.ParamsRange = ast.CloneRange(other.ParamsRange)
	}

	return diagnostics.EmptyDiags()
}

func overrideSentinelOptions(this, other *ast.File, overrideValues bool) diagnostics.Diagnostics {
	diags := diagnostics.EmptyDiags()

	if this == nil {
		return diags
	}

	if !overrideValues || other.SentinelOptions == nil {
		return diags
	}

	if this.SentinelOptions == nil {
		this.SentinelOptions = ast.CloneSentinelOptions(other.SentinelOptions)
		return diags
	}

	if other.SentinelOptions.FeaturesRange != nil {
		this.SentinelOptions.FeaturesRange = ast.CloneRange(other.SentinelOptions.FeaturesRange)
		this.SentinelOptions.Features = ast.CloneFeatureList(other.SentinelOptions.Features)
	}

	return diags
}

// Import Overriders
func overrideImports(this, other *ast.File, overrideValues bool) diagnostics.Diagnostics {
	diags := diagnostics.EmptyDiags()

	if this == nil || other == nil {
		return diags
	}

	for name, otherNode := range other.Imports {
		if otherNode == nil {
			continue
		}
		if _, ok := this.Imports[name]; ok {
			diags = append(diags, overrideImport(this.Imports[name], other.Imports[name], overrideValues)...)
		} else {
			diags = diags.Add(&diagnostics.Diagnostic{
				Severity: diagnostics.Error,
				Summary:  fmt.Sprintf("Missing base %s declaration to override", otherNode.BlockType()),
				Detail: fmt.Sprintf("There is no %s named %q. An override file can only override a %s that was already declared in a primary configuration file.",
					otherNode.BlockType(),
					otherNode.BlockName(),
					otherNode.BlockType(),
				),
				Range: ast.CloneRange(otherNode.Range()),
			})
		}
	}

	return diags
}

func overrideImport(this, other ast.Import, overrideValues bool) diagnostics.Diagnostics {
	if this == nil || other == nil {
		return diagnostics.EmptyDiags()
	}

	switch actual := this.(type) {
	case *ast.V1ModuleImport:
		return overrideV1ModuleImportWith(actual, other, overrideValues)
	case *ast.V1PluginImport:
		return overrideV1PluginImportWith(actual, other, overrideValues)
	case *ast.V2ModuleImport:
		return overrideV2ModuleImportWith(actual, other, overrideValues)
	case *ast.V2PluginImport:
		return overrideV2PluginImportWith(actual, other, overrideValues)
	case *ast.V2StaticImport:
		return overrideV2StaticImportWith(actual, other, overrideValues)
	default:
		panic(fmt.Sprintf("Unknown import type %T", actual))
	}
}

func unexpectedImportTypeDiag(this, that ast.Import) diagnostics.Diagnostics {
	return diagnostics.Diagnostics{
		{
			Severity: diagnostics.Error,
			Summary:  `Unexpected import type`,
			Detail: fmt.Sprintf("Cannot override a %s with a %s",
				this.BlockType(),
				that.BlockType(),
			),
			Range: ast.CloneRange(that.Range()),
		},
	}
}

func overrideV1ModuleImportWith(this *ast.V1ModuleImport, imp ast.Import, overrideValues bool) diagnostics.Diagnostics {
	if other, ok := imp.(*ast.V1ModuleImport); !ok {
		return unexpectedImportTypeDiag(this, imp)
	} else {
		if other.SourceRange != nil && overrideValues {
			this.Source = other.Source
			this.SourceRange = ast.CloneRange(other.SourceRange)
		}
	}

	return diagnostics.EmptyDiags()
}

func overrideV1PluginImportWith(this *ast.V1PluginImport, imp ast.Import, overrideValues bool) diagnostics.Diagnostics {
	if other, ok := imp.(*ast.V1PluginImport); !ok {
		return unexpectedImportTypeDiag(this, imp)
	} else {
		if other.ArgsRange != nil && overrideValues {
			this.Args = ast.CloneStringList(other.Args)
			this.ArgsRange = ast.CloneRange(other.ArgsRange)
		}
		if other.ConfigRange != nil && overrideValues {
			this.Config = ast.CloneParameterMap(other.Config)
			this.ConfigRange = ast.CloneRange(other.ConfigRange)
		}
		if other.EnvRange != nil && overrideValues {
			this.Env = ast.CloneStringList(other.Env)
			this.EnvRange = ast.CloneRange(other.EnvRange)
		}
		if other.PathRange != nil && overrideValues {
			this.Path = other.Path
			this.PathRange = ast.CloneRange(other.PathRange)
		}
	}

	return diagnostics.EmptyDiags()
}

func overrideV2ModuleImportWith(this *ast.V2ModuleImport, imp ast.Import, overrideValues bool) diagnostics.Diagnostics {
	if other, ok := imp.(*ast.V2ModuleImport); !ok {
		return unexpectedImportTypeDiag(this, imp)
	} else {
		if other.SourceRange != nil && overrideValues {
			this.Source = other.Source
			this.SourceRange = ast.CloneRange(other.SourceRange)
		}
	}

	return diagnostics.EmptyDiags()
}

func overrideV2PluginImportWith(this *ast.V2PluginImport, imp ast.Import, overrideValues bool) diagnostics.Diagnostics {
	if other, ok := imp.(*ast.V2PluginImport); !ok {
		return unexpectedImportTypeDiag(this, imp)
	} else {
		if other.ArgsRange != nil && overrideValues {
			this.Args = ast.CloneStringList(other.Args)
			this.ArgsRange = ast.CloneRange(other.ArgsRange)
		}
		if other.ConfigRange != nil && overrideValues {
			this.Config = ast.CloneParameterMap(other.Config)
			this.ConfigRange = ast.CloneRange(other.ConfigRange)
		}
		if other.EnvRange != nil && overrideValues {
			this.Env = ast.CloneParameterMap(other.Env)
			this.EnvRange = ast.CloneRange(other.EnvRange)
		}
		if other.SourceRange != nil && overrideValues {
			this.Source = other.Source
			this.SourceRange = ast.CloneRange(other.SourceRange)
		}
	}

	return diagnostics.EmptyDiags()
}

func overrideV2StaticImportWith(this *ast.V2StaticImport, imp ast.Import, overrideValues bool) diagnostics.Diagnostics {
	if other, ok := imp.(*ast.V2StaticImport); !ok {
		return unexpectedImportTypeDiag(this, imp)
	} else {
		if other.SourceRange != nil && overrideValues {
			this.Source = other.Source
			this.SourceRange = ast.CloneRange(other.SourceRange)
		}
		if other.FormatRange != nil && overrideValues {
			this.Format = other.Format
			this.FormatRange = ast.CloneRange(other.FormatRange)
		}
	}

	return diagnostics.EmptyDiags()
}

func overrideMap[E ast.HCLNode](
	thisMap map[string]*E,
	otherMap map[string]*E,
	overrideFunc func(key string) diagnostics.Diagnostics,
) diagnostics.Diagnostics {

	diags := diagnostics.EmptyDiags()
	for name, otherNode := range otherMap {
		if otherNode == nil {
			continue
		}
		if _, ok := thisMap[name]; ok {
			diags = diags.Concat(overrideFunc(name))
		} else {
			diags = diags.Add(&diagnostics.Diagnostic{
				Severity: diagnostics.Error,
				Summary:  fmt.Sprintf("Missing base %s declaration to override", (*otherNode).BlockType()),
				Detail: fmt.Sprintf("There is no %s named %q. An override file can only override a %s that was already declared in a primary configuration file.",
					(*otherNode).BlockType(),
					(*otherNode).BlockName(),
					(*otherNode).BlockType(),
				),
				Range: ast.CloneRange((*otherNode).Range()),
			})
		}
	}

	return diags
}
