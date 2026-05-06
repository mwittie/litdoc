package bash

import "litdoc/internal"

type Runner struct{}

func (r Runner) Run(script string) (string, string, int, error) {
	return internal.RunCmd("bash", "-c", script)
}
