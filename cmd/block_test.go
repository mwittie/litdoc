package cmd_test

import "testing"

func TestBlockCommand(t *testing.T) {
	runScript(t, "testdata/script/block.txtar")
}
