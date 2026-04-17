package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var blockOutput string

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Process a block",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(string(data))
	},
}

func init() {
	blockCmd.Flags().StringVarP(&blockOutput, "output", "o", "", "output format")
	rootCmd.AddCommand(blockCmd)
}
