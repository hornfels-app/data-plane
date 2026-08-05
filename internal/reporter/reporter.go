package reporter

import (
	"encoding/json"
	"fmt"
	"os"
)

// Violation represents a single column that failed the Hornfels check.
type Violation struct {
	Table       string `json:"table"`
	Column      string `json:"column"`
	DataType    string `json:"data_type"`
	Reason      string `json:"reason"`
	ProposedFix string `json:"proposed_fix"`
}

// Receipt represents the overall output state.
type Receipt struct {
	Status     string      `json:"status"`
	Violations []Violation `json:"violations"`
}

// WriteJSONReceipt writes the compliance receipt to disk.
func WriteJSONReceipt(receipt *Receipt, path string) error {
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// PrintStdout prints a human-readable summary to the terminal.
func PrintStdout(receipt *Receipt) {
	if receipt.Status == "PASS" {
		fmt.Println("✅ Hornfels Check Passed. Zero unclassified PII columns found.")
		return
	}

	fmt.Printf("❌ Hornfels Check Failed! Found %d unclassified columns.\n\n", len(receipt.Violations))
	for _, v := range receipt.Violations {
		fmt.Printf("Table:  %s\nColumn: %s\nReason: %s\nFix:\n  %s\n\n", v.Table, v.Column, v.Reason, v.ProposedFix)
	}
}
