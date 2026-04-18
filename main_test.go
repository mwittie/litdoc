package main_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestScript(t *testing.T) {
	binPath, err := filepath.Abs("bin")
	if err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Setup: func(env *testscript.Env) error {
			env.Setenv("PATH", binPath+string(os.PathListSeparator)+env.Getenv("PATH"))
			return nil
		},
	})
}
