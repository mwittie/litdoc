package internal

import (
	"fmt"
	"os"
)

func ProcessFile(srcFilePath string) (string, error) {
	srcContent, err := os.ReadFile(srcFilePath)
	if err != nil {
		return "", fmt.Errorf("reading source file: %w", err)
	}

	blocks, err := MakeBlocksFromMarkdown(srcContent)
	if err != nil {
		return "", fmt.Errorf("parsing source file: %w", err)
	}

	cells, err := Classify(blocks)
	if err != nil {
		return "", fmt.Errorf("classifying blocks into cells: %w", err)
	}

	cells, err = Execute(cells)
	if err != nil {
		return "", fmt.Errorf("executing cells: %w", err)
	}

	dstContent, err := Compose(cells)
	if err != nil {
		return "", fmt.Errorf("composing cells into content: %w", err)
	}

	return dstContent, nil
}
