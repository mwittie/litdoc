package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type Runner interface {
	Run(script string) (stdout, stderr string, exitCode int, err error)
}

func RunCmd(name string, args ...string) (string, string, int, error) {
	var outBuf, errBuf bytes.Buffer
	var exitCode int

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return "", "", 0, fmt.Errorf("running command: %w", err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode, nil
}
