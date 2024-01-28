package json

import (
	"bytes"
	"flag"
	"os"
	"path"
	"strings"
	"testing"

	"io"

	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/glennsarti/sentinel-parser/sentinel_config/parser"
	"github.com/glennsarti/sentinel-parser/sentinel_config/serialization"
	jsonSerializer "github.com/glennsarti/sentinel-parser/sentinel_config/serialization/json"
	"github.com/glennsarti/sentinel-parser/sentinel_config/serialization/tracing"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

const archiveInputFile = "sentinel.hcl"
const archiveDiagOutput = "diagOut.txt"

var debugSentinel = flag.Bool("sentinel-debug", false, "Turn on verbose debugging of Sentinel parsing")

type serializersUnderTest map[string]serialization.Serializer

func TestSentinelConfigSerializers(t *testing.T) {
	fixturesDir := path.Join("..", "..", "spec", "test-fixtures")

	var tr tracing.Tracer
	if *debugSentinel {
		tr = NewSpecTracer(t)
	}

	// List the serializers to test here
	sut := serializersUnderTest{
		"json": jsonSerializer.New(tr),
	}

	items, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range items {
		if item.IsDir() {
			t.Run(item.Name(), func(t *testing.T) {
				processTestFixturesDir(item.Name(), fixturesDir, item.Name(), sut, t)
			})
		}
	}
}

func NewSpecTracer(t *testing.T) tracing.Tracer {
	return &specTracer{t: t}
}

type specTracer struct {
	t      *testing.T
	indent int
}

func (st *specTracer) PushMethod(method, message string) {
	st.t.Logf("%s(%s) BEGIN %s",
		strings.Repeat("| ", st.indent),
		method,
		message,
	)
	st.indent += 1
}
func (st *specTracer) Trace(message string) {
	st.t.Logf("%s%s",
		strings.Repeat("| ", st.indent),
		message,
	)
}
func (st *specTracer) PopMethod(message string) {
	st.indent -= 1
	st.t.Logf("%s(%s) END %s",
		strings.Repeat("| ", st.indent),
		"???",
		message,
	)
}

func processTestFixturesDir(relPath, srcDir, sentinelVersion string, sut serializersUnderTest, t *testing.T) {
	dirPath := path.Join(srcDir, relPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	// To test serializers, we take all of the test fixtures for parsing, and then round-trip
	// serialize -> deserialize and ensure it generates the same AST. This means we can only assert
	// on examples that parse without error.
	//
	// If a serializer _really_ needs to test specific examples
	// they should exist within the serializers directory.

	for _, entry := range entries {
		if entry.IsDir() {
			t.Run(entry.Name(), func(t *testing.T) {
				processTestFixturesDir(path.Join(relPath, entry.Name()), srcDir, sentinelVersion, sut, t)
			})
		} else {
			if strings.HasSuffix(entry.Name(), ".txtar") {
				if err := testSpecFile(entry.Name(), dirPath, sentinelVersion, sut, t); err != nil {
					t.Error(err)
				}
			}
		}
	}
}

func testSpecFile(filename, parentPath, sentinelVersion string, sut serializersUnderTest, t *testing.T) error {
	filePath := path.Join(parentPath, filename)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	var arc *txtar.Archive = nil
	var inputFile txtar.File
	var diagFile txtar.File
	hasDiags := false

	if contents, err := io.ReadAll(f); err != nil {
		return err
	} else {
		if err := f.Close(); err != nil {
			return err
		}

		arc = txtar.Parse(contents)
		for _, f := range arc.Files {
			switch f.Name {
			case archiveInputFile:
				inputFile = f
			case archiveDiagOutput:
				hasDiags = true
				diagFile = f
			}
		}
	}

	// Has it been parsed before and is fine?
	if !hasDiags || len(diagFile.Data) > 0 {
		return nil
	}

	t.Run(filename, func(t *testing.T) {
		// Create the parser
		hclParser, err := parser.NewParser(sentinelVersion)
		if err != nil {
			t.Error(err)
			return
		}

		// Parse it
		expectedAst, diags := hclParser.ParseFileSource(inputFile.Data, archiveInputFile)
		if diags.HasErrors() {
			t.Error(diags)
			return
		}

		for name, subject := range sut {
			t.Run(name, func(t *testing.T) {
				// Serialize it
				buf := bytes.NewBufferString("")
				err = subject.Serialize(expectedAst, buf)
				if err != nil {
					t.Error(err)
					return
				}

				// Deserialize it
				actualAst := ast.NewFile()
				err = subject.Deserialize(buf, actualAst)
				if err != nil {
					t.Error(err)
					return
				}

				if diff := cmp.Diff(expectedAst, actualAst, diffCompareOptions()); diff != "" {
					t.Fatal(diff)
				}
			})
		}
	})
	return nil
}

func diffCompareOptions() cmp.Options {
	return cmp.Options{
		cmp.Comparer(ast.DynamicValueComparer),
		cmp.Comparer(ast.DynamicValuePtrComparer),
	}
}
