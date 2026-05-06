package internal_test

import (
	_ "embed"
	"os"
	"testing"

	"litdoc/internal"
	"litdoc/internal/bash"
	"litdoc/internal/static"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/input.md
var renderInput []byte

//go:embed testdata/output.md
var renderOutput []byte

func TestProcessFile(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"input to output", renderInput, renderOutput},
		{"output to output", renderOutput, renderOutput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			f, err := os.CreateTemp(t.TempDir(), "*.md")
			require.NoError(t, err)
			_, err = f.Write(tt.input)
			require.NoError(t, err)
			err = f.Close()
			require.NoError(t, err)

			// when
			// todo: pull these definitions to given
			got, err := internal.ProcessFile(
				f.Name(), map[string]internal.CellParser{"static": internal.CellParserFunc(static.ParseStaticCell), "bash": internal.CellParserFunc(bash.ParseBashCell)})

			// then
			require.NoError(t, err)
			assert.Equal(t, string(tt.want), got)
		})
	}
}

func TestProcessFileNestedListIndent(t *testing.T) {
	// given
	input := "- Level 1\n\n  - Level 2\n\n    ```bash | litdoc\n    echo hello\n    ```\n"
	f, err := os.CreateTemp(t.TempDir(), "*.md")
	require.NoError(t, err)
	_, err = f.Write([]byte(input))
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// when
	got, err := internal.ProcessFile(f.Name(), map[string]internal.CellParser{"static": internal.CellParserFunc(static.ParseStaticCell), "bash": internal.CellParserFunc(bash.ParseBashCell)})

	// then
	require.NoError(t, err)
	want := "- Level 1\n\n  - Level 2\n\n    ```bash | litdoc\n    echo hello\n    ```\n" +
		"\n    " + internal.OutputBeginMarker + "\n" +
		"    output\n" +
		"    " + internal.OutputEndMarker + "\n"
	assert.Equal(t, want, got)
}
