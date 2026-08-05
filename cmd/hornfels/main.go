package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hornfels",
	Short: "Hornfels - The Developer's PII Control Layer",
	Long: `Hornfels blocks PII schema leaks in CI/CD by enforcing strict 
COMMENT ON COLUMN tagging policies.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
