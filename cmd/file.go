package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"litdoc/internal"

	"github.com/spf13/cobra"
)

var fileWrite bool

var fileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Process a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		data, err := internal.ProcessFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if fileWrite {
			if err := writeFileAtomic(path, data); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
		fmt.Print(data)
	},
}

func writeFileAtomic(path, data string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".litdoc-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func init() {
	fileCmd.Flags().BoolVarP(&fileWrite, "write", "w", false, "rewrite the file in place")
	rootCmd.AddCommand(fileCmd)
}
