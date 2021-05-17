package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "domain",
	Short: "An esoteric quantum information repository",
	Long:  "Domain is a CLI application to manage your digital library",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(serveCmd)
    rootCmd.AddCommand(loadCmd)
    rootCmd.AddCommand(saveCmd)
    rootCmd.AddCommand(addCmd)
}
