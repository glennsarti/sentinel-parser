package spec

import (
	"io"
	"os"

	"golang.org/x/tools/txtar"
)

const archiveInputFile = "input.sentinel"
const archiveAstOutput = "astOut.txt"
const archiveErrorOutput = "astError.txt"
const archiveTokensOutput = "tokens.txt"

type parsedArchive struct {
	InputFile  txtar.File
	AstOutFile txtar.File
	ErrorsFile txtar.File
	TokensFile txtar.File
	raw        *txtar.Archive
}

func parseTxtarArchive(filePath string) (*parsedArchive, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	arc := &parsedArchive{}
	arc.raw = txtar.Parse(contents)

	for _, f := range arc.raw.Files {
		switch f.Name {
		case archiveInputFile:
			arc.InputFile = f
		case archiveAstOutput:
			arc.AstOutFile = f
		case archiveErrorOutput:
			arc.ErrorsFile = f
		case archiveTokensOutput:
			arc.TokensFile = f
		}
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
		defer f.Close()
		if _, err = f.Write(txtar.Format(pa.raw)); err != nil {
			return err
		}
	}
	return nil
}
