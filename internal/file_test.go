package internal_test

import (
	_ "embed"
	"os"
	"testing"

	"litdoc/internal"
)

//go:embed testdata/input.md
var renderInput []byte

//go:embed testdata/output.md
var renderOutput []byte

func TestProcessFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(renderInput); err != nil {
		t.Fatal(err)
	}
	f.Close()

	got, err := internal.ProcessFile(f.Name())
	if err != nil {
		t.Fatalf("ProcessFile: %v", err)
	}

	if got != string(renderOutput) {
		t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, renderOutput)
	}
}
