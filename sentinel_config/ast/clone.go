package ast

import (
	"fmt"

	"github.com/glennsarti/sentinel-parser/position"
)

func asPointer[E any](item E) *E {
	return &item
}

func CloneFile(f *File) *File {
	if f == nil {
		return nil
	}
	dst := File{}

	dst.Globals = CloneGlobalMap(f.Globals)
	dst.Imports = CloneImportMap(f.Imports)
	dst.Mocks = CloneMockMap(f.Mocks)
	dst.Params = CloneParameterMap(f.Params)
	dst.Policies = ClonePolicyMap(f.Policies)
	dst.SentinelOptions = CloneSentinelOptions(f.SentinelOptions)
	dst.Test = CloneTest(f.Test)

	return &dst
}

func CloneStringList(src []string) []string {
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func CloneRange(src *position.SourceRange) *position.SourceRange {
	if src == nil {
		return nil
	}
	result := src.Clone()
	return &result
}

func CloneDynamicValue(src *DynamicValue) *DynamicValue {
	if src == nil {
		return nil
	}
	clone := src.Clone()
	return &clone
}

func CloneFeature(src Feature) Feature {
	return Feature{
		Name:       src.Name,
		NameRange:  CloneRange(src.NameRange),
		Value:      src.Value,
		ValueType:  src.ValueType,
		ValueRange: CloneRange(src.ValueRange),
	}
}

func CloneFeatureList(src []*Feature) []*Feature {
	dst := make([]*Feature, len(src))
	for idx, item := range src {
		if item != nil {
			dst[idx] = asPointer(CloneFeature(*item))
		} else {
			dst[idx] = nil
		}
	}
	return src
}

func CloneGlobal(src Global) Global {
	return Global{
		Name:        src.Name,
		NameRange:   CloneRange(src.NameRange),
		Value:       CloneDynamicValue(src.Value),
		ValueType:   src.ValueType,
		ValueRange:  CloneRange(src.ValueRange),
		GlobalRange: CloneRange(src.GlobalRange),
	}
}

func CloneGlobalMap(src map[string]*Global) map[string]*Global {
	dst := make(map[string]*Global, 0)
	for name, item := range src {
		if item != nil {
			dst[name] = asPointer(CloneGlobal(*item))
		} else {
			dst[name] = nil
		}
	}
	return dst
}

func CloneImport(src Import) Import {
	switch actual := src.(type) {
	case *V1ModuleImport:
		return CloneV1ModuleImport(actual)
	case *V1PluginImport:
		return CloneV1PluginImport(actual)
	case *V2ModuleImport:
		return CloneV2ModuleImport(actual)
	case *V2PluginImport:
		return CloneV2PluginImport(actual)
	case *V2StaticImport:
		return CloneV2StaticImport(actual)
	default:
		panic(fmt.Sprintf("Unknown import type %T", actual))
	}
}

func CloneImportMap(src map[string]Import) map[string]Import {
	dst := make(map[string]Import, 0)
	for name, item := range src {
		dst[name] = CloneImport(item)
	}
	return dst
}

func CloneMock(src Mock) Mock {
	dst := Mock{
		Name:      src.Name,
		NameRange: CloneRange(src.NameRange),
		DataRange: CloneRange(src.DataRange),
		MockRange: CloneRange(src.MockRange),
	}

	dst.Data = CloneParameterMap(src.Data)
	dst.Module = CloneMockModule(src.Module)

	return dst
}

func CloneMockMap(src map[string]*Mock) map[string]*Mock {
	dst := make(map[string]*Mock, 0)
	for name, item := range src {
		if item != nil {
			dst[name] = asPointer(CloneMock(*item))
		} else {
			dst[name] = nil
		}
	}
	return dst
}

func CloneMockModule(src *MockModule) *MockModule {
	if src == nil {
		return nil
	}
	return &MockModule{
		Source:          src.Source,
		SourceRange:     CloneRange(src.SourceRange),
		MockModuleRange: CloneRange(src.MockModuleRange),
	}
}

func CloneParameter(src Parameter) Parameter {
	return Parameter{
		Name:           src.Name,
		NameRange:      CloneRange(src.NameRange),
		Value:          CloneDynamicValue(src.Value),
		ValueType:      src.ValueType,
		ValueRange:     CloneRange(src.ValueRange),
		ParameterRange: CloneRange(src.ParameterRange),
	}
}

func CloneParameterMap(src map[string]*Parameter) map[string]*Parameter {
	dst := make(map[string]*Parameter, 0)
	for name, item := range src {
		if item != nil {
			dst[name] = asPointer(CloneParameter(*item))
		} else {
			dst[name] = nil
		}
	}
	return dst
}

func ClonePolicy(src Policy) Policy {
	dst := Policy{
		Name:                  src.Name,
		NameRange:             CloneRange(src.NameRange),
		Source:                src.Source,
		SourceRange:           CloneRange(src.SourceRange),
		EnforcementLevel:      src.EnforcementLevel,
		EnforcementLevelRange: CloneRange(src.EnforcementLevelRange),
		ParamsRange:           CloneRange(src.ParamsRange),
		PolicyRange:           CloneRange(src.PolicyRange),
	}

	dst.Params = CloneParameterMap(src.Params)

	return dst
}

func ClonePolicyMap(src map[string]*Policy) map[string]*Policy {
	dst := make(map[string]*Policy, 0)
	for name, item := range src {
		if item != nil {
			dst[name] = asPointer(ClonePolicy(*item))
		} else {
			dst[name] = nil
		}
	}
	return dst
}

func CloneSentinelOptions(src *SentinelOptions) *SentinelOptions {
	if src == nil {
		return nil
	}
	dst := SentinelOptions{
		SentinelOptionsRange: CloneRange(src.SentinelOptionsRange),
		FeaturesRange:        CloneRange(src.FeaturesRange),
	}

	dst.Features = CloneFeatureList(src.Features)

	return &dst
}

func CloneTest(src *Test) *Test {
	if src == nil {
		return nil
	}
	dst := Test{
		RulesRange: CloneRange(src.RulesRange),
		TestRange:  CloneRange(src.TestRange),
	}

	if src.Rules != nil {
		dst.Rules = make([]*TestRule, len(src.Rules))
		for idx := range src.Rules {
			dst.Rules[idx] = CloneTestRule(src.Rules[idx])
		}
	}

	return &dst
}

func CloneTestRule(src *TestRule) *TestRule {
	if src == nil {
		return nil
	}
	return &TestRule{
		TestRuleRange: CloneRange(src.TestRuleRange),
		Name:          src.Name,
		NameRange:     CloneRange(src.NameRange),
		ValueType:     src.ValueType,
		ValueRange:    CloneRange(src.ValueRange),
	}
}

func CloneV1ModuleImport(src *V1ModuleImport) Import {
	return &V1ModuleImport{
		Name:        src.Name,
		NameRange:   CloneRange(src.NameRange),
		Source:      src.Source,
		SourceRange: CloneRange(src.SourceRange),
		BlockRange:  CloneRange(src.BlockRange),
	}
}

func CloneV1PluginImport(src *V1PluginImport) Import {
	dst := V1PluginImport{
		Name:        src.Name,
		NameRange:   CloneRange(src.NameRange),
		Path:        src.Path,
		PathRange:   CloneRange(src.PathRange),
		ArgsRange:   CloneRange(src.ArgsRange),
		EnvRange:    CloneRange(src.EnvRange),
		ConfigRange: CloneRange(src.ConfigRange),
		BlockRange:  CloneRange(src.BlockRange),
	}

	if src.Args != nil {
		dst.Args = CloneStringList(src.Args)
	}

	dst.Config = CloneParameterMap(src.Config)

	if src.Env != nil {
		dst.Env = CloneStringList(src.Env)
	}

	return &dst
}

func CloneV2ModuleImport(src *V2ModuleImport) Import {
	return &V2ModuleImport{
		Kind:        src.Kind,
		KindRange:   CloneRange(src.KindRange),
		Name:        src.Name,
		NameRange:   CloneRange(src.NameRange),
		Source:      src.Source,
		SourceRange: CloneRange(src.SourceRange),
		BlockRange:  CloneRange(src.BlockRange),
	}
}

func CloneV2PluginImport(src *V2PluginImport) Import {
	dst := V2PluginImport{
		Name:        src.Name,
		NameRange:   CloneRange(src.NameRange),
		Source:      src.Source,
		SourceRange: CloneRange(src.SourceRange),
		ArgsRange:   CloneRange(src.ArgsRange),
		Kind:        src.Kind,
		KindRange:   CloneRange(src.KindRange),
		EnvRange:    CloneRange(src.EnvRange),
		ConfigRange: CloneRange(src.ConfigRange),
		BlockRange:  CloneRange(src.BlockRange),
	}

	if src.Args != nil {
		dst.Args = CloneStringList(src.Args)
	}

	dst.Config = CloneParameterMap(src.Config)
	dst.Env = CloneParameterMap(src.Env)

	return &dst
}

func CloneV2StaticImport(src *V2StaticImport) Import {
	return &V2StaticImport{
		Name:        src.Name,
		NameRange:   CloneRange(src.NameRange),
		Kind:        src.Kind,
		KindRange:   CloneRange(src.KindRange),
		Format:      src.Format,
		FormatRange: CloneRange(src.FormatRange),
		Source:      src.Source,
		SourceRange: CloneRange(src.SourceRange),
		BlockRange:  CloneRange(src.BlockRange),
	}
}
