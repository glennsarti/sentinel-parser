package parser

// import (
// 	"fmt"
// 	"sort"
// 	"testing"

// 	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
// 	"github.com/google/go-cmp/cmp"
// 	"github.com/hashicorp/hcl/v2"
// 	"github.com/zclconf/go-cty/cty"
// )

// var fixturesDir = "../spec/test-fixtures"

// func diffCompareOptions() cmp.Options {
// 	return cmp.Options{
// 		cmp.Comparer(ctyValueComparer),
// 		cmp.Comparer(ctyValuePtrComparer),
// 		cmp.Comparer(dynamicValueComparer),
// 		cmp.Comparer(dynamicValuePtrComparer),
// 	}
// }

// // Needed?
// // func HRange(sline, scol, eline, ecol int) *hcl.Range {
// // 	return hRange(sline, scol, eline, ecol)
// // }

// func hRange(sline, scol, eline, ecol int) *ast.SourceRange {
// 	val := ast.SourceRange{
// 		Filename: "",
// 		Start: ast.SourcePos{
// 			Line:   sline,
// 			Column: scol,
// 		},
// 		End: ast.SourcePos{
// 			Line:   eline,
// 			Column: ecol,
// 		},
// 	}
// 	return &val
// }

// func hFileRange(filename string, sline, scol, eline, ecol int) *ast.SourceRange {
// 	val := hRange(sline, scol, eline, ecol)
// 	val.Filename = filename
// 	return val
// }

// func asValuePtr(v cty.Value) *cty.Value {
// 	return &v
// }

// func ctyValueComparer(x, y cty.Value) bool {
// 	return x.GoString() == y.GoString()
// }

// func ctyValuePtrComparer(x, y *cty.Value) bool {
// 	if x == nil && y == nil {
// 		return true
// 	}
// 	if x != nil && y != nil {
// 		return x.GoString() == y.GoString()
// 	}
// 	return false
// }

// func dynamicValueComparer(x, y ast.DynamicValue) bool {
// 	return x.GoString() == y.GoString()
// }

// func dynamicValuePtrComparer(x, y *ast.DynamicValue) bool {
// 	if x == nil && y == nil {
// 		return true
// 	}
// 	if x != nil && y != nil {
// 		return x.GoString() == y.GoString()
// 	}
// 	return false
// }

// func rangeToString(r *hcl.Range) string {
// 	if r == nil {
// 		return ":nil"
// 	}
// 	return fmt.Sprintf(":%d,%d->%d,%d", r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
// }

// func compareDiagnostics(t *testing.T, expectedDiags []string, actual Diagnostics, sev hcl.DiagnosticSeverity) {
// 	actualDiags := make([]string, 0)

// 	for _, diag := range actual {
// 		if diag.Severity == sev {
// 			relatedString := ""
// 			if diag.Related != nil {
// 				for _, r := range *diag.Related {
// 					relatedString = relatedString + "\n" + r.Summary + rangeToString(r.Subject)
// 				}
// 			}
// 			errMsg := diag.Detail + rangeToString(diag.Subject) + relatedString
// 			actualDiags = append(actualDiags, errMsg)
// 		}
// 	}

// 	if len(actualDiags) == 0 && expectedDiags == nil {
// 		return
// 	}

// 	sort.Strings(expectedDiags)
// 	sort.Strings(actualDiags)

// 	if diff := cmp.Diff(expectedDiags, actualDiags); diff != "" {
// 		t.Fatal(diff)
// 	}
// }

// func stripRangeInfo(old *ast.SourceRange, removeEntirely bool, removeFilename bool) *ast.SourceRange {
// 	if old == nil || removeEntirely {
// 		return nil
// 	}

// 	filename := old.Filename
// 	if removeFilename {
// 		filename = ""
// 	}

// 	return &ast.SourceRange{
// 		Filename: filename,
// 		Start: ast.SourcePos{
// 			Line:   old.Start.Line,
// 			Column: old.Start.Column,
// 		},
// 		End: ast.SourcePos{
// 			Line:   old.End.Line,
// 			Column: old.End.Column,
// 		},
// 	}
// }

// func stripRangeInfoInFile(file *ast.File) {
// 	doStripRangeInfoInFile(file, false, true)
// }

// func removeRangeInfoInFile(file *ast.File) {
// 	doStripRangeInfoInFile(file, true, false)
// }

// func stripByteOffsetInRangeInfoInFile(file *ast.File) {
// 	doStripRangeInfoInFile(file, false, false)
// }

// func doStripRangeInfoInFile(file *ast.File, removeEntirely bool, removeFilename bool) {
// 	if file == nil {
// 		return
// 	}

// 	for idx := range file.Globals {
// 		file.Globals[idx].NameRange = stripRangeInfo(file.Globals[idx].NameRange, removeEntirely, removeFilename)
// 		file.Globals[idx].ValueRange = stripRangeInfo(file.Globals[idx].ValueRange, removeEntirely, removeFilename)
// 		file.Globals[idx].GlobalRange = stripRangeInfo(file.Globals[idx].GlobalRange, removeEntirely, removeFilename)
// 	}

// 	for idx := range file.Imports {
// 		switch actual := file.Imports[idx].(type) {
// 		case *ast.V1ModuleImport:
// 			actual.NameRange = stripRangeInfo(actual.NameRange, removeEntirely, removeFilename)
// 			actual.BlockRange = stripRangeInfo(actual.BlockRange, removeEntirely, removeFilename)
// 			actual.SourceRange = stripRangeInfo(actual.SourceRange, removeEntirely, removeFilename)
// 		case *ast.V1PluginImport:
// 			actual.ArgsRange = stripRangeInfo(actual.ArgsRange, removeEntirely, removeFilename)
// 			actual.ConfigRange = stripRangeInfo(actual.ConfigRange, removeEntirely, removeFilename)
// 			actual.EnvRange = stripRangeInfo(actual.EnvRange, removeEntirely, removeFilename)
// 			actual.BlockRange = stripRangeInfo(actual.BlockRange, removeEntirely, removeFilename)
// 			actual.NameRange = stripRangeInfo(actual.NameRange, removeEntirely, removeFilename)
// 			actual.PathRange = stripRangeInfo(actual.PathRange, removeEntirely, removeFilename)
// 			for idx := range actual.Config {
// 				actual.Config[idx].NameRange = stripRangeInfo(actual.Config[idx].NameRange, removeEntirely, removeFilename)
// 				actual.Config[idx].ValueRange = stripRangeInfo(actual.Config[idx].ValueRange, removeEntirely, removeFilename)
// 				actual.Config[idx].ParameterRange = stripRangeInfo(actual.Config[idx].ParameterRange, removeEntirely, removeFilename)
// 			}
// 		case *ast.V2ModuleImport:
// 			actual.KindRange = stripRangeInfo(actual.KindRange, removeEntirely, removeFilename)
// 			actual.NameRange = stripRangeInfo(actual.NameRange, removeEntirely, removeFilename)
// 			actual.BlockRange = stripRangeInfo(actual.BlockRange, removeEntirely, removeFilename)
// 			actual.SourceRange = stripRangeInfo(actual.SourceRange, removeEntirely, removeFilename)
// 		case *ast.V2PluginImport:
// 			actual.ArgsRange = stripRangeInfo(actual.ArgsRange, removeEntirely, removeFilename)
// 			actual.ConfigRange = stripRangeInfo(actual.ConfigRange, removeEntirely, removeFilename)
// 			actual.EnvRange = stripRangeInfo(actual.EnvRange, removeEntirely, removeFilename)
// 			actual.KindRange = stripRangeInfo(actual.KindRange, removeEntirely, removeFilename)
// 			actual.BlockRange = stripRangeInfo(actual.BlockRange, removeEntirely, removeFilename)
// 			actual.NameRange = stripRangeInfo(actual.NameRange, removeEntirely, removeFilename)
// 			actual.SourceRange = stripRangeInfo(actual.SourceRange, removeEntirely, removeFilename)
// 			for idx := range actual.Env {
// 				actual.Env[idx].NameRange = stripRangeInfo(actual.Env[idx].NameRange, removeEntirely, removeFilename)
// 				actual.Env[idx].ValueRange = stripRangeInfo(actual.Env[idx].ValueRange, removeEntirely, removeFilename)
// 				actual.Env[idx].ParameterRange = stripRangeInfo(actual.Env[idx].ParameterRange, removeEntirely, removeFilename)
// 			}
// 			for idx := range actual.Config {
// 				actual.Config[idx].NameRange = stripRangeInfo(actual.Config[idx].NameRange, removeEntirely, removeFilename)
// 				actual.Config[idx].ValueRange = stripRangeInfo(actual.Config[idx].ValueRange, removeEntirely, removeFilename)
// 				actual.Config[idx].ParameterRange = stripRangeInfo(actual.Config[idx].ParameterRange, removeEntirely, removeFilename)
// 			}
// 		case *ast.V2StaticImport:
// 			actual.FormatRange = stripRangeInfo(actual.FormatRange, removeEntirely, removeFilename)
// 			actual.KindRange = stripRangeInfo(actual.KindRange, removeEntirely, removeFilename)
// 			actual.NameRange = stripRangeInfo(actual.NameRange, removeEntirely, removeFilename)
// 			actual.BlockRange = stripRangeInfo(actual.BlockRange, removeEntirely, removeFilename)
// 			actual.SourceRange = stripRangeInfo(actual.SourceRange, removeEntirely, removeFilename)
// 		}
// 	}

// 	for idx := range file.Mocks {
// 		file.Mocks[idx].NameRange = stripRangeInfo(file.Mocks[idx].NameRange, removeEntirely, removeFilename)
// 		file.Mocks[idx].DataRange = stripRangeInfo(file.Mocks[idx].DataRange, removeEntirely, removeFilename)
// 		file.Mocks[idx].MockRange = stripRangeInfo(file.Mocks[idx].MockRange, removeEntirely, removeFilename)

// 		for didx := range file.Mocks[idx].Data {
// 			file.Mocks[idx].Data[didx].NameRange = stripRangeInfo(file.Mocks[idx].Data[didx].NameRange, removeEntirely, removeFilename)
// 			file.Mocks[idx].Data[didx].ValueRange = stripRangeInfo(file.Mocks[idx].Data[didx].ValueRange, removeEntirely, removeFilename)
// 		}

// 		if file.Mocks[idx].Module != nil {
// 			file.Mocks[idx].Module.SourceRange = stripRangeInfo(file.Mocks[idx].Module.SourceRange, removeEntirely, removeFilename)
// 			file.Mocks[idx].Module.MockModuleRange = stripRangeInfo(file.Mocks[idx].Module.MockModuleRange, removeEntirely, removeFilename)
// 		}
// 	}

// 	if file.SentinelOptions != nil {
// 		file.SentinelOptions.FeaturesRange = stripRangeInfo(file.SentinelOptions.FeaturesRange, removeEntirely, removeFilename)
// 		file.SentinelOptions.SentinelOptionsRange = stripRangeInfo(file.SentinelOptions.SentinelOptionsRange, removeEntirely, removeFilename)
// 		for idx := range file.SentinelOptions.Features {
// 			file.SentinelOptions.Features[idx].NameRange = stripRangeInfo(file.SentinelOptions.Features[idx].NameRange, removeEntirely, removeFilename)
// 			file.SentinelOptions.Features[idx].ValueRange = stripRangeInfo(file.SentinelOptions.Features[idx].ValueRange, removeEntirely, removeFilename)
// 		}
// 	}

// 	for idx := range file.Params {
// 		file.Params[idx].NameRange = stripRangeInfo(file.Params[idx].NameRange, removeEntirely, removeFilename)
// 		file.Params[idx].ValueRange = stripRangeInfo(file.Params[idx].ValueRange, removeEntirely, removeFilename)
// 		file.Params[idx].ParameterRange = stripRangeInfo(file.Params[idx].ParameterRange, removeEntirely, removeFilename)
// 	}

// 	for idx := range file.Policies {
// 		file.Policies[idx].EnforcementLevelRange = stripRangeInfo(file.Policies[idx].EnforcementLevelRange, removeEntirely, removeFilename)
// 		file.Policies[idx].NameRange = stripRangeInfo(file.Policies[idx].NameRange, removeEntirely, removeFilename)
// 		file.Policies[idx].SourceRange = stripRangeInfo(file.Policies[idx].SourceRange, removeEntirely, removeFilename)
// 		file.Policies[idx].ParamsRange = stripRangeInfo(file.Policies[idx].ParamsRange, removeEntirely, removeFilename)
// 		file.Policies[idx].PolicyRange = stripRangeInfo(file.Policies[idx].PolicyRange, removeEntirely, removeFilename)

// 		for pidx := range file.Policies[idx].Params {
// 			file.Policies[idx].Params[pidx].NameRange = stripRangeInfo(file.Policies[idx].Params[pidx].NameRange, removeEntirely, removeFilename)
// 			file.Policies[idx].Params[pidx].ValueRange = stripRangeInfo(file.Policies[idx].Params[pidx].ValueRange, removeEntirely, removeFilename)
// 			file.Policies[idx].Params[pidx].ParameterRange = stripRangeInfo(file.Policies[idx].Params[pidx].ParameterRange, removeEntirely, removeFilename)
// 		}
// 	}

// 	if file.Test != nil {
// 		file.Test.RulesRange = stripRangeInfo(file.Test.RulesRange, removeEntirely, removeFilename)
// 		file.Test.TestRange = stripRangeInfo(file.Test.TestRange, removeEntirely, removeFilename)

// 		for ridx := range file.Test.Rules {
// 			file.Test.Rules[ridx].NameRange = stripRangeInfo(file.Test.Rules[ridx].NameRange, removeEntirely, removeFilename)
// 			file.Test.Rules[ridx].ValueRange = stripRangeInfo(file.Test.Rules[ridx].ValueRange, removeEntirely, removeFilename)
// 			file.Test.Rules[ridx].TestRuleRange = stripRangeInfo(file.Test.Rules[ridx].TestRuleRange, removeEntirely, removeFilename)
// 		}
// 	}
// }
