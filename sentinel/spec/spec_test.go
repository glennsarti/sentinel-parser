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
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/lexer"
	"github.com/glennsarti/sentinel-parser/sentinel/parser"
	scjson "github.com/glennsarti/sentinel-parser/sentinel/serialization/json"
	"github.com/glennsarti/sentinel-parser/sentinel/token"
	"github.com/glennsarti/sentinel-parser/sentinel/tracing"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

var debugSentinel = flag.Bool("sentinel-debug", false, "Turn on verbose debugging of Sentinel parsing")
var updateTestFiles = flag.Bool("update-test-files", false, "Update the language test files")

func TestSentinelLanguageSpecs(t *testing.T) {
	fixturesDir := path.Join("test-fixtures")

	items, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range items {
		if item.IsDir() {
			if valid, actualVersion := features.ValidateSentinelVersion(item.Name()); valid {
				t.Run(item.Name(), func(t *testing.T) {
					processTestFixturesDir(item.Name(), fixturesDir, actualVersion, t)
				})
			} else {
				t.Logf("Invalid directory name for a sentinel version: '%s'", item.Name())
			}
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
			if strings.HasSuffix(entry.Name(), ".sentinel") {
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
	dstFilename := strings.Replace(filename, ".sentinel", ".txtar", -1)
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

	if arc.InputFile.Name == "" {
		return fmt.Errorf("Missing input file in archive %s", filePath)
	}

	tokens, err := generateTokens(sentinelVersion, arc.InputFile.Data)
	if err != nil {
		return err
	}
	arc.UpdateFile(archiveTokensOutput, tokens)

	// Parse it
	result, err := generateAST(sentinelVersion, arc.InputFile.Name, arc.InputFile.Data, t)
	if err != nil {
		return err
	}

	arc.UpdateFile(archiveAstOutput, result.Output)
	arc.UpdateFile(archiveErrorOutput, result.Diags)

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

	actualTokens, err := generateTokens(sentinelVersion, arc.InputFile.Data)
	if err != nil {
		return err
	}

	t.Run("tokens", func(t *testing.T) {
		if diff := cmp.Diff(string(arc.TokensFile.Data), string(actualTokens)); diff != "" {
			debugOutputStrings(string(arc.TokensFile.Data), string(actualTokens), "Tokens", t)
			t.Fatal(diff)
		}
	})

	actual, err := generateAST(sentinelVersion, arc.InputFile.Name, arc.InputFile.Data, t)
	if err != nil {
		return err
	}

	t.Run("ast", func(t *testing.T) {
		expectedString := string(arc.AstOutFile.Data)
		actualString := string(actual.Output)
		if diff := cmp.Diff(expectedString, actualString); diff != "" {
			debugOutputStrings(expectedString, actualString, "AST Output", t)
			t.Fatal(diff)
		}
	})

	t.Run("errors", func(t *testing.T) {
		expectedString := string(arc.ErrorsFile.Data)
		actualString := string(actual.Diags)
		if diff := cmp.Diff(expectedString, actualString); diff != "" {
			debugOutputStrings(expectedString, actualString, "AST Errors", t)
			t.Fatal(diff)
		}
	})

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

func generateTokens(sentinelVersion string, source []byte) ([]byte, error) {
	lex, err := lexer.New(sentinelVersion, string(source))
	if err != nil {
		return []byte{}, err
	}

	output := ""
	for {
		pos, endPos, tok := lex.NextToken()

		lit := strings.ReplaceAll(tok.Literal, "\n", "\\n")
		lit = strings.ReplaceAll(lit, "\r", "")
		lit = strings.ReplaceAll(lit, "\t", "\\t")

		output = output + fmt.Sprintf("%d:%d:%s:%s", pos, endPos, tok.Type, lit) + "\n"

		if tok.Type == token.EOF {
			break
		}
	}

	return []byte(output), nil
}

type astResult struct {
	Output []byte
	Diags  []byte
	Ast    *ast.File
}

func generateAST(sentinelVersion, filename string, source []byte, t *testing.T) (*astResult, error) {
	var tracer tracing.ParsingTracer
	if *debugSentinel {
		tracer = NewSpecTracer(t)
	}

	fast, _, diags, err := parser.ParseFileWithTracing(sentinelVersion, filename, source, tracer)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString("")
	if err := scjson.New(nil).Serialize(fast, buf); err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, buf.Bytes(), "", "  "); err != nil {
		return nil, err
	}

	astOut := prettyJSON.Bytes()

	// Must always end in an LF otherwise it doesn't diff correctly
	astOut = append(astOut, '\n')

	return &astResult{
		Output: astOut,
		Ast:    fast,
		Diags:  diagnosticsToBytes(diags),
	}, nil
}

func diagnosticsToBytes(diags diagnostics.Diagnostics) []byte {
	diagStrs := make([]string, 0)

	for _, diag := range diags.Errors() {
		errMsg := "ERROR: "
		errMsg = errMsg + diag.Detail + rangeToString(diag.Range)
		diagStrs = append(diagStrs, errMsg)
	}
	sort.Strings(diagStrs)

	var out bytes.Buffer
	for _, val := range diagStrs {
		out.WriteString(val + "\n")
	}

	return out.Bytes()
}

func rangeToString(r *position.SourceRange) string {
	if r == nil {
		return ":null"
	}
	return fmt.Sprintf(":%d,%d->%d,%d", r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
}
