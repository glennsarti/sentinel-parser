package ast

// Autogenerated using generate_sentinel_config_ast.js
// DO NOT MODIFY MANUALLY

import (
	"github.com/glennsarti/sentinel-parser/position"
)

type Feature struct {
	Name       string
	NameRange  *position.SourceRange
	Value      *DynamicValue
	ValueRange *position.SourceRange
	ValueType  string
}

type File struct {
	Globals         map[string]*Global
	Imports         map[string]Import
	Mocks           map[string]*Mock
	Params          map[string]*Parameter
	Policies        map[string]*Policy
	SentinelOptions *SentinelOptions
	Test            *Test
}

type Global struct {
	GlobalRange *position.SourceRange
	Name        string
	NameRange   *position.SourceRange
	Value       *DynamicValue
	ValueRange  *position.SourceRange
	ValueType   string
}

type Mock struct {
	Data      map[string]*Parameter
	DataRange *position.SourceRange
	MockRange *position.SourceRange
	Module    *MockModule
	Name      string
	NameRange *position.SourceRange
}

type MockModule struct {
	MockModuleRange *position.SourceRange
	Source          string
	SourceRange     *position.SourceRange
}

type Parameter struct {
	Name           string
	NameRange      *position.SourceRange
	ParameterRange *position.SourceRange
	Value          *DynamicValue
	ValueRange     *position.SourceRange
	ValueType      string
}

type Policy struct {
	EnforcementLevel      string
	EnforcementLevelRange *position.SourceRange
	Name                  string
	NameRange             *position.SourceRange
	Params                map[string]*Parameter
	ParamsRange           *position.SourceRange
	PolicyRange           *position.SourceRange
	Source                string
	SourceRange           *position.SourceRange
}

type SentinelOptions struct {
	Features             []*Feature
	FeaturesRange        *position.SourceRange
	SentinelOptionsRange *position.SourceRange
}

type Test struct {
	Rules      []*TestRule
	RulesRange *position.SourceRange
	TestRange  *position.SourceRange
}

type TestRule struct {
	Name          string
	NameRange     *position.SourceRange
	TestRuleRange *position.SourceRange
	ValueRange    *position.SourceRange
	ValueType     string
}

type V1ModuleImport struct {
	BlockRange  *position.SourceRange
	Name        string
	NameRange   *position.SourceRange
	Source      string
	SourceRange *position.SourceRange
}

type V1PluginImport struct {
	Args        []string
	ArgsRange   *position.SourceRange
	BlockRange  *position.SourceRange
	Config      map[string]*Parameter
	ConfigRange *position.SourceRange
	Env         []string
	EnvRange    *position.SourceRange
	Name        string
	NameRange   *position.SourceRange
	Path        string
	PathRange   *position.SourceRange
}

type V2ModuleImport struct {
	BlockRange  *position.SourceRange
	Kind        string
	KindRange   *position.SourceRange
	Name        string
	NameRange   *position.SourceRange
	Source      string
	SourceRange *position.SourceRange
}

type V2PluginImport struct {
	Args        []string
	ArgsRange   *position.SourceRange
	BlockRange  *position.SourceRange
	Config      map[string]*Parameter
	ConfigRange *position.SourceRange
	Env         map[string]*Parameter
	EnvRange    *position.SourceRange
	Kind        string
	KindRange   *position.SourceRange
	Name        string
	NameRange   *position.SourceRange
	Source      string
	SourceRange *position.SourceRange
}

type V2StaticImport struct {
	BlockRange  *position.SourceRange
	Format      string
	FormatRange *position.SourceRange
	Kind        string
	KindRange   *position.SourceRange
	Name        string
	NameRange   *position.SourceRange
	Source      string
	SourceRange *position.SourceRange
}

// All nodes must implement Range()
func (o Feature) Range() *position.SourceRange         { return nil }
func (o File) Range() *position.SourceRange            { return nil }
func (o Global) Range() *position.SourceRange          { return o.GlobalRange }
func (o Mock) Range() *position.SourceRange            { return o.MockRange }
func (o MockModule) Range() *position.SourceRange      { return nil }
func (o Parameter) Range() *position.SourceRange       { return o.ParameterRange }
func (o Policy) Range() *position.SourceRange          { return o.PolicyRange }
func (o SentinelOptions) Range() *position.SourceRange { return o.SentinelOptionsRange }
func (o Test) Range() *position.SourceRange            { return o.TestRange }
func (o TestRule) Range() *position.SourceRange        { return nil }
func (o V1ModuleImport) Range() *position.SourceRange  { return o.BlockRange }
func (o V1PluginImport) Range() *position.SourceRange  { return o.BlockRange }
func (o V2ModuleImport) Range() *position.SourceRange  { return o.BlockRange }
func (o V2PluginImport) Range() *position.SourceRange  { return o.BlockRange }
func (o V2StaticImport) Range() *position.SourceRange  { return o.BlockRange }

// HCLNode implementations
func (o Global) BlockName() string                              { return o.Name }
func (o Global) BlockNameRange() *position.SourceRange          { return o.NameRange }
func (o Global) BlockType() string                              { return "global" }
func (o Mock) BlockName() string                                { return o.Name }
func (o Mock) BlockNameRange() *position.SourceRange            { return o.NameRange }
func (o Mock) BlockType() string                                { return "mock" }
func (o Parameter) BlockName() string                           { return o.Name }
func (o Parameter) BlockNameRange() *position.SourceRange       { return o.NameRange }
func (o Parameter) BlockType() string                           { return "param" }
func (o Policy) BlockName() string                              { return o.Name }
func (o Policy) BlockNameRange() *position.SourceRange          { return o.NameRange }
func (o Policy) BlockType() string                              { return "policy" }
func (o SentinelOptions) BlockName() string                     { return "sentinel" }
func (o SentinelOptions) BlockNameRange() *position.SourceRange { return nil }
func (o SentinelOptions) BlockType() string                     { return "sentinel" }
func (o Test) BlockName() string                                { return "test" }
func (o Test) BlockNameRange() *position.SourceRange            { return nil }
func (o Test) BlockType() string                                { return "test" }
func (o V1ModuleImport) BlockName() string                      { return o.Name }
func (o V1ModuleImport) BlockNameRange() *position.SourceRange  { return o.NameRange }
func (o V1ModuleImport) BlockType() string                      { return "module" }
func (o V1PluginImport) BlockName() string                      { return o.Name }
func (o V1PluginImport) BlockNameRange() *position.SourceRange  { return o.NameRange }
func (o V1PluginImport) BlockType() string                      { return "import" }
func (o V2ModuleImport) BlockName() string                      { return o.Name }
func (o V2ModuleImport) BlockNameRange() *position.SourceRange  { return o.NameRange }
func (o V2ModuleImport) BlockType() string                      { return "module import" }
func (o V2PluginImport) BlockName() string                      { return o.Name }
func (o V2PluginImport) BlockNameRange() *position.SourceRange  { return o.NameRange }
func (o V2PluginImport) BlockType() string                      { return "plugin import" }
func (o V2StaticImport) BlockName() string                      { return o.Name }
func (o V2StaticImport) BlockNameRange() *position.SourceRange  { return o.NameRange }
func (o V2StaticImport) BlockType() string                      { return "static import" }

// Import implementations
func (i *V1ModuleImport) Schema() ImportSchema { return V1ImportSchema }
func (i *V1PluginImport) Schema() ImportSchema { return V1ImportSchema }
func (i *V2ModuleImport) Schema() ImportSchema { return V2ImportSchema }
func (i *V2PluginImport) Schema() ImportSchema { return V2ImportSchema }
func (i *V2StaticImport) Schema() ImportSchema { return V2ImportSchema }
