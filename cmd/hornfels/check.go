package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hornfels-app/data-plane/internal/heuristics"
	"github.com/hornfels-app/data-plane/internal/policy"
	"github.com/hornfels-app/data-plane/internal/reporter"
	"github.com/hornfels-app/data-plane/internal/scanner"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the database schema against the Hornfels policy",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := policy.LoadConfig(".hornfels.yaml")
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		base, err := policy.LoadBaseline(".hornfels-baseline.yaml")
		if err != nil {
			fmt.Println("Error loading baseline:", err)
			os.Exit(1)
		}

		usePrisma, _ := cmd.Flags().GetBool("prisma")

		var sc scanner.Scanner
		ctx := context.Background()

		if usePrisma {
			sc = scanner.NewPrismaScanner("schema.prisma")
		} else {
			dbURL := os.Getenv("DATABASE_URL")
			if dbURL == "" {
				fmt.Println("DATABASE_URL environment variable is required when not using --prisma")
				os.Exit(1)
			}
			sc, err = scanner.NewScannerFromURL(ctx, dbURL)
			if err != nil {
				fmt.Println("Failed to connect to database:", err)
				os.Exit(1)
			}
		}
		defer sc.Close()

		cols, err := sc.ScanSchema(ctx)
		if err != nil {
			fmt.Println("Failed to scan schema:", err)
			os.Exit(1)
		}

		scanData, _ := cmd.Flags().GetBool("scan-data")

		rcpt := &reporter.Receipt{Status: "PASS"}
		
		for _, col := range cols {
			// Skip excluded tables (simple logic for now)
			excluded := false
			for _, ex := range cfg.Exclude {
				if col.Table == ex {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}

			// Skip baseline ignored
			if base.IsIgnored(col.Table, col.Name) {
				continue
			}

			hasTag, isPII := scanner.HasHornfelsTag(col.Comment)

			// Heuristic Check
			isSuspicious := heuristics.IsSuspiciousColumn(col.Name)
			
			if cfg.StrictMode && !hasTag {
				reason := "Column missing [hornfels: pii=true|false] classification."
				if isSuspicious {
					reason = "Column name looks like PII but is unclassified."
				}
				
				fix := fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '[hornfels: pii=true]';", col.Table, col.Name)
				rcpt.Violations = append(rcpt.Violations, reporter.Violation{
					Table: col.Table, Column: col.Name, DataType: col.DataType,
					Reason: reason, ProposedFix: fix,
				})
			} else if scanData && hasTag && !isPII {
				// They marked it as non-PII, let's sample data to verify
				samples, _ := sc.SampleData(ctx, col.Table)
				foundPII := false
				for _, row := range samples {
					val, ok := row[col.Name].(string)
					if ok {
						if heuristics.ContainsEmail(val) || heuristics.ContainsSSN(val) || heuristics.IsValidLuhn(val) {
							foundPII = true
							break
						}
					}
				}
				if foundPII {
					rcpt.Violations = append(rcpt.Violations, reporter.Violation{
						Table: col.Table, Column: col.Name, DataType: col.DataType,
						Reason: "Column tagged as pii=false but sampled data contains SSN/Email/CreditCard.",
						ProposedFix: fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '[hornfels: pii=true]';", col.Table, col.Name),
					})
				}
			}
		}

		if len(rcpt.Violations) > 0 {
			rcpt.Status = "FAIL"
		}

		reporter.PrintStdout(rcpt)
		reporter.WriteJSONReceipt(rcpt, "hornfels-receipt.json")
		reporter.PostGitHubPRComment(ctx, rcpt)

		if rcpt.Status == "FAIL" && cfg.Enforce {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	
	// Flags
	checkCmd.Flags().Bool("scan-data", false, "Sample data to detect unstructured PII")
	checkCmd.Flags().Bool("prisma", false, "Parse schema.prisma instead of querying the live DB")
}
