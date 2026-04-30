package cmd_test

import "testing"

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