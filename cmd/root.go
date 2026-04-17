package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var output string

var rootCmd = &cobra.Command{
	Use:   "litdoc",
	Short: "Literate documentation tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("output:", output)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "output format")
}
