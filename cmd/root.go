package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "litdoc",
	Short: "Literate documentation tool",
}

func Execute() error {
	return rootCmd.Execute()
}
