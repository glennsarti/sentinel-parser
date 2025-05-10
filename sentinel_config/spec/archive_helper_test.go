package spec

import (
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/txtar"
)

const archiveInputFile = "sentinel.hcl"
const archiveAstOutput = "astOut.json"
const archiveDiagOutput = "diagOut.txt"
const archiveOverrideFile = "override.hcl"

type parsedArchive struct {
	PrimaryFile    txtar.File
	AstOutFile     txtar.File
	DiagnosticFile txtar.File
	OverrideFiles  []txtar.File
	raw            *txtar.Archive
}

func parseTxtarArchive(filePath string) (*parsedArchive, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	arc := &parsedArchive{}
	arc.raw = txtar.Parse(contents)

	overrideFiles := map[string]txtar.File{}
	overrideNames := []string{}

	for _, f := range arc.raw.Files {
		switch f.Name {
		case archiveInputFile:
			arc.PrimaryFile = f
		case archiveAstOutput:
			arc.AstOutFile = f
		case archiveDiagOutput:
			arc.DiagnosticFile = f
		case archiveOverrideFile:
			overrideFiles[archiveOverrideFile] = f
			overrideNames = append(overrideNames, f.Name)
		}
		if strings.HasSuffix(f.Name, "_override.hcl") {
			overrideFiles[f.Name] = f
			overrideNames = append(overrideNames, f.Name)
		}
	}

	sort.Strings(overrideNames)
	arc.OverrideFiles = make([]txtar.File, len(overrideNames))
	for idx, name := range overrideNames {
		arc.OverrideFiles[idx] = overrideFiles[name]
	}

	return arc, nil
}

func (pa *parsedArchive) UpdateFile(name string, content []byte) {
	fileIdx := -1
	for idx, f := range pa.raw.Files {
		if f.Name == name {
			fileIdx = idx
			break
		}
	}

	tf := txtar.File{
		Name: name,
		Data: content,
	}

	if fileIdx == -1 {
		pa.raw.Files = append(pa.raw.Files, tf)
	} else {
		pa.raw.Files[fileIdx] = tf
	}
}

func (pa *parsedArchive) Write(filepath string) error {
	if f, err := os.Create(filepath); err != nil {
		return err
	} else {
		defer f.Close() //nolint:errcheck
		if _, err = f.Write(txtar.Format(pa.raw)); err != nil {
			return err
		}
	}
	return nil
}
