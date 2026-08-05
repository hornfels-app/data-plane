package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Hornfels configuration in the current repository",
	Run: func(cmd *cobra.Command, args []string) {
		cfgData := "enforce: true\nstrict_mode: true\nexclude: []\n"
		if err := os.WriteFile(".hornfels.yaml", []byte(cfgData), 0644); err != nil {
			fmt.Println("Error writing config:", err)
			return
		}

		cursorRules := `---
description: Strict constraints to prevent hallucinations and bloat for coding agents working on Hornfels.
alwaysApply: true
---
# Hornfels Rules
Do not guess schemas. Use hornfels tags.
`
		if err := os.WriteFile(".cursorrules", []byte(cursorRules), 0644); err != nil {
			fmt.Println("Error writing .cursorrules:", err)
			return
		}

		fmt.Println("✅ Hornfels initialized. .hornfels.yaml and .cursorrules created.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
