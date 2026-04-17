package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var file string

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render a file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("file:", file)
	},
}

func init() {
	renderCmd.Flags().StringVarP(&file, "file", "f", "", "file to render")
	renderCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(renderCmd)
}
