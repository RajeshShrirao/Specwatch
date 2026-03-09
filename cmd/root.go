package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	versionOk bool
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
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("specwatch version %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  date: %s\n", date)
		},
	})
}

func findSpecFile() string {
	paths := []string{"spec.md", "./spec.md", "../spec.md"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
