package json

import (
	"bytes"
	"flag"
	"os"
	"path"
	"strings"
	"testing"

	"io"

	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/sentinel/ast"
	"github.com/glennsarti/sentinel-parser/sentinel/parser"
	"github.com/glennsarti/sentinel-parser/sentinel/serialization"
	jsonSerializer "github.com/glennsarti/sentinel-parser/sentinel/serialization/json"
	"github.com/glennsarti/sentinel-parser/sentinel/serialization/tracing"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

const archiveInputFile = "input.sentinel"
const archiveAstErrors = "astError.txt"

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
			if valid, actualVersion := features.ValidateSentinelVersion(item.Name()); valid {
				t.Run(item.Name(), func(t *testing.T) {
					processTestFixturesDir(item.Name(), fixturesDir, actualVersion, sut, t)
				})
			} else {
				t.Logf("Invalid directory name for a sentinel version: '%s'", item.Name())
			}
		}
	}
}

func processTestFixturesDir(relPath, srcDir, sentinelVersion string, sut serializersUnderTest, t *testing.T) {
	dirPath := path.Join(srcDir, relPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

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
	var errorFile txtar.File
	hasErrors := false

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
			case archiveAstErrors:
				hasErrors = true
				errorFile = f
			}
		}
	}

	// Has it been parsed before and is fine?
	if !hasErrors || len(errorFile.Data) > 0 {
		return nil
	}

	t.Run(filename, func(t *testing.T) {
		// Create the parser
		expectedAst, _, _, err := parser.ParseFile(sentinelVersion, inputFile.Name, inputFile.Data)
		if err != nil {
			t.Error(err)
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

				if diff := cmp.Diff(expectedAst, actualAst); diff != "" {
					//t.Fatal("DIFF")
					t.Fatal(diff)
				}
			})
		}
	})
	return nil
}
