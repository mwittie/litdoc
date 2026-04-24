package cmd

import (
	"fmt"
	"os"

	"litdoc/internal"

	"github.com/spf13/cobra"
)

var inputFile string

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Process a file",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := internal.ProcessFile(inputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	fileCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file to process")
	fileCmd.MarkFlagRequired("input")
	rootCmd.AddCommand(fileCmd)
}
