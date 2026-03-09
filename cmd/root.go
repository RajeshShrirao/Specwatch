package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "specwatch",
	Short: "specwatch is a tool for watching and analyzing specs",
	Long:  `A fast, structured spec-driven static analysis tool for modern web development.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to specwatch! Use 'specwatch help' for more information.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be defined here
}
