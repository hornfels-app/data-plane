package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hornfels-app/data-plane/internal/policy"
	"github.com/hornfels-app/data-plane/internal/scanner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Generate .hornfels-baseline.yaml by ignoring existing schema",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			fmt.Println("DATABASE_URL environment variable is required")
			os.Exit(1)
		}

		ctx := context.Background()
		sc, err := scanner.NewScannerFromURL(ctx, dbURL)
		if err != nil {
			fmt.Println("Failed to connect to database:", err)
			os.Exit(1)
		}
		defer sc.Close()

		cols, err := sc.ScanSchema(ctx)
		if err != nil {
			fmt.Println("Failed to scan schema:", err)
			os.Exit(1)
		}

		b := &policy.Baseline{
			IgnoredColumns: make(map[string][]string),
		}

		for _, col := range cols {
			b.IgnoredColumns[col.Table] = append(b.IgnoredColumns[col.Table], col.Name)
		}

		data, err := yaml.Marshal(b)
		if err != nil {
			fmt.Println("Failed to marshal baseline:", err)
			os.Exit(1)
		}

		if err := os.WriteFile(".hornfels-baseline.yaml", data, 0644); err != nil {
			fmt.Println("Failed to write baseline file:", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Baseline generated successfully. Ignored %d columns.\n", len(cols))
	},
}

func init() {
	rootCmd.AddCommand(baselineCmd)
}
