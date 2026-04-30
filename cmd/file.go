package cmd

import (
	"fmt"
	"os"

	"litdoc/internal"

	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Process a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := internal.ProcessFile(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(data)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
}
