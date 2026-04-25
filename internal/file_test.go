package internal_test

import (
	_ "embed"
	"os"
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestProcessFileNestedListIndent(t *testing.T) {
	// given
	input := "- Level 1\n\n  - Level 2\n\n    ```bash | litdoc\n    echo hello\n    ```\n"
	f, err := os.CreateTemp(t.TempDir(), "*.md")
	require.NoError(t, err)
	_, err = f.Write([]byte(input))
	require.NoError(t, err)
	f.Close()

	// when
	got, err := internal.ProcessFile(f.Name())

	// then
	require.NoError(t, err)
	want := "- Level 1\n\n  - Level 2\n\n    ```bash | litdoc\n    echo hello\n    ```\n" +
		"\n    " + internal.OutputBeginMarker +
		"    output\n" +
		"    " + internal.OutputEndMarker
	assert.Equal(t, want, got)
}
