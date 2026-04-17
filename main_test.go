package main_test

import (
	"os"
	"testing"

	"litdoc/cmd"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"litdoc": func() {
			if err := cmd.Execute(); err != nil {
				os.Exit(1)
			}
		},
	})
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
