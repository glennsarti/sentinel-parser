package spec

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/glennsarti/sentinel-parser/sentinel_config/parser"
	scjson "github.com/glennsarti/sentinel-parser/sentinel_config/serialization/json"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

var debugSentinel = flag.Bool("sentinel-debug", false, "Turn on verbose debugging of Sentinel parsing")
var updateTestFiles = flag.Bool("update-test-files", false, "Update the language test files")

func TestSentinelConfigLanguageSpecs(t *testing.T) {
	fixturesDir := path.Join("test-fixtures")

	items, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range items {
		if item.IsDir() {
			t.Run(item.Name(), func(t *testing.T) {
				processTestFixturesDir(item.Name(), fixturesDir, item.Name(), t)
			})
		}
	}
}

func processTestFixturesDir(relPath, srcDir, sentinelVersion string, t *testing.T) {
	dirPath := path.Join(srcDir, relPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Run(entry.Name(), func(t *testing.T) {
				processTestFixturesDir(path.Join(relPath, entry.Name()), srcDir, sentinelVersion, t)
			})
		} else {
			if strings.HasSuffix(entry.Name(), ".hcl") {
				t.Run(entry.Name(), func(t *testing.T) {
					if err := createSpecFileFromSentinel(entry.Name(), dirPath, sentinelVersion, t); err != nil {
						t.Error(err)
					}
				})
			}
			if strings.HasSuffix(entry.Name(), ".txtar") {
				t.Run(entry.Name(), func(t *testing.T) {
					if *updateTestFiles {
						if err := updateSpecFile(entry.Name(), dirPath, sentinelVersion, t); err != nil {
							t.Error(err)
						}
					} else {
						if err := testSpecFile(entry.Name(), dirPath, sentinelVersion, t); err != nil {
							t.Error(err)
						}
					}
				})
			}
		}
	}
}

func createSpecFileFromSentinel(filename, parentPath, sentinelVersion string, t *testing.T) error {
	srcFile := path.Join(parentPath, filename)
	dstFilename := strings.Replace(filename, ".hcl", ".txtar", -1)
	dstFile := path.Join(parentPath, dstFilename)

	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	contents, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	f.Close()

	// Must end in LF
	if contents[len(contents)-1] != 10 {
		contents = append(contents, 10)
	}

	arc := &txtar.Archive{}
	arc.Files = append(arc.Files, txtar.File{
		Name: archiveInputFile,
		Data: contents,
	})

	if f, err = os.Create(dstFile); err != nil {
		return err
	} else {
		if _, err = f.Write(txtar.Format(arc)); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	if err := updateSpecFile(dstFilename, parentPath, sentinelVersion, t); err != nil {
		return err
	}

	if err := os.Remove(srcFile); err != nil {
		return err
	}

	return nil
}

func updateSpecFile(filename, parentPath, sentinelVersion string, t *testing.T) error {
	filePath := path.Join(parentPath, filename)

	arc, err := parseTxtarArchive(filePath)
	if err != nil {
		return err
	}

	if arc.PrimaryFile.Name == "" {
		return fmt.Errorf("Missing input file in archive %s", filePath)
	}

	// Parse it
	result, err := generateAST(sentinelVersion, arc.PrimaryFile, arc.OverrideFiles)
	if err != nil {
		return err
	}
	// Must always end in an LF otherwise it doesn't diff correctly
	result.Output = append(result.Output, '\n')

	arc.UpdateFile(archiveAstOutput, result.Output)
	arc.UpdateFile(archiveDiagOutput, result.Diags)

	if err := arc.Write(filePath); err != nil {
		return err
	}

	t.Logf("Updated %s", filePath)
	return err
}

func testSpecFile(filename, parentPath, sentinelVersion string, t *testing.T) error {
	filePath := path.Join(parentPath, filename)

	arc, err := parseTxtarArchive(filePath)
	if err != nil {
		return err
	}

	if arc.PrimaryFile.Name == "" {
		return fmt.Errorf("Missing input file in archive %s", filePath)
	}

	if arc.AstOutFile.Name == "" || arc.DiagnosticFile.Name == "" {
		return updateSpecFile(filename, parentPath, sentinelVersion, t)
	}

	actual, err := generateAST(sentinelVersion, arc.PrimaryFile, arc.OverrideFiles)
	if err != nil {
		return fmt.Errorf("Fatal error parsing %s", arc.PrimaryFile.Name)
	}

	// Must always end in an LF otherwise it doesn't diff correctly
	actual.Output = append(actual.Output, '\n')

	t.Run("ast", func(t *testing.T) {
		expectedString := string(arc.AstOutFile.Data)
		actualString := string(actual.Output)
		if diff := cmp.Diff(expectedString, actualString); diff != "" {
			debugOutputStrings(expectedString, actualString, "AST Output", t)
			t.Fatal(diff)
		}
	})

	t.Run("errors", func(t *testing.T) {
		expectedString := string(arc.DiagnosticFile.Data)
		actualString := string(actual.Diags)
		if diff := cmp.Diff(expectedString, actualString); diff != "" {
			debugOutputStrings(expectedString, actualString, "AST Errors", t)
			t.Fatal(diff)
		}
	})

	t.Run("clone", func(t *testing.T) {
		clone := ast.CloneFile(actual.Ast)
		if diff := cmp.Diff(actual.Ast, clone, diffCompareOptions()); diff != "" {
			t.Fatal(diff)
		}
	})

	// Can only assert here if:
	// * We have no overrides
	// * We actually have an AST to work with
	// * We have a supported sentinel version
	if len(arc.OverrideFiles) == 0 &&
		actual.Ast != nil &&
		features.SupportedVersion(sentinelVersion, features.ConfigurationOverrideMinimumVersion) {
		t.Run("attempt_override", func(t *testing.T) {
			otherData := []byte("# Modified Source\n")
			otherData = append(otherData, arc.PrimaryFile.Data...)
			otherName := "other_" + arc.PrimaryFile.Name
			other, err := generateAST(sentinelVersion, txtar.File{Name: otherName, Data: otherData}, []txtar.File{})
			if err != nil {
				t.Fatalf("error parsing %s", otherName)
			}
			// Clone the AST because we don't want to modify it
			clone := ast.CloneFile(actual.Ast)

			// Attempt to override
			if err := parser.AttemptOverrideFileWith(clone, other.Ast, sentinelVersion); err != nil && err.HasErrors() {
				t.Fatalf("error attempting override %s", err)
			}

			// No changes should occur when we attempt to override
			if diff := cmp.Diff(actual.Ast, clone, diffCompareOptions()); diff != "" {
				t.Fatal(diff)
			}
		})
	}

	return nil
}

func debugOutputStrings(expected, actual, thingName string, t *testing.T) {
	if !*debugSentinel {
		return
	}
	t.Logf("--- Begin Expected %s", thingName)
	t.Log(string(expected))
	t.Logf("--- End Expected %s\n", thingName)

	t.Logf("--- Begin Actual %s", thingName)
	t.Log(string(actual))
	t.Logf("--- End Actual %s", thingName)
}

type astResult struct {
	Output []byte
	Diags  []byte
	Ast    *ast.File
}

func generateAST(sentinelVersion string, primary txtar.File, overrides []txtar.File) (*astResult, error) {
	subject, err := parser.NewParser(sentinelVersion)
	if err != nil {
		return nil, err
	}

	ast, diags := subject.ParseFileSource(primary.Data, primary.Name)

	for _, f := range overrides {
		otherAst, otherDiags := subject.ParseFileSource(f.Data, f.Name)
		diags = append(diags, otherDiags...)
		otherDiags = parser.OverrideFileWith(ast, otherAst, sentinelVersion)
		diags = append(diags, otherDiags...)
	}

	buf := bytes.NewBufferString("")
	if err := scjson.New(nil).Serialize(ast, buf); err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, buf.Bytes(), "", "  "); err != nil {
		return nil, err
	}

	result := astResult{
		Output: prettyJSON.Bytes(),
		Diags:  diagnosticsToBytes(diags),
		Ast:    ast,
	}

	return &result, nil
}

func diagnosticsToBytes(diags diagnostics.Diagnostics) []byte {
	diagStrs := make([]string, 0)

	for _, diag := range diags {
		errMsg := "ERROR: "
		switch diag.Severity {
		case diagnostics.Unknown:
			errMsg = "INVALID: "
		case diagnostics.Warning:
			errMsg = "WARNING: "
		}
		errMsg = errMsg + diag.Detail + rangeToString(diag.Range)

		if diag.Related != nil && len(*diag.Related) > 0 {
			relatedStrings := make([]string, 0)
			for _, r := range *diag.Related {
				relatedStrings = append(relatedStrings, fmt.Sprintf("  RELATED: %s%s", r.Summary, rangeToString(r.Range)))
			}
			sort.Strings(relatedStrings)

			errMsg = errMsg + "\n" + strings.Join(relatedStrings, "\n")
		}
		diagStrs = append(diagStrs, errMsg)
	}
	sort.Strings(diagStrs)

	var out bytes.Buffer
	for _, val := range diagStrs {
		out.WriteString(val + "\n")
	}

	return out.Bytes()
}
