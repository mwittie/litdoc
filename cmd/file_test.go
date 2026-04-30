package cmd_test

import (
	"fmt"
	"os"
	"testing"

	"litdoc/cmd"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"litdoc": func() {
			if err := cmd.Execute(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	})
}

func TestFileCommand(t *testing.T) {
	runScript(t, "testdata/script/file.txtar")
}

func TestFileWriteFlag(t *testing.T) {
	runScript(t, "testdata/script/file_write.txtar")
}

func TestFileCommandNoArgs(t *testing.T) {
	runScript(t, "testdata/script/file_no_args.txtar")
}

func TestFileCommandMissingPath(t *testing.T) {
	runScript(t, "testdata/script/file_missing_path.txtar")
}

func runScript(t *testing.T, script string) {
	t.Helper()

	testscript.Run(t, testscript.Params{
		Files: []string{script},
	})
}
