package heuristics

import (
	"regexp"
)

var (
	// Schema Regex: Detects likely PII columns based on their names.
	ssnColumnRegex   = regexp.MustCompile(`(?i)(ssn|social_security|socialsecurity)`)
	emailColumnRegex = regexp.MustCompile(`(?i)(email)`)
	phoneColumnRegex = regexp.MustCompile(`(?i)(phone|mobile|cell|telephone)`)
	ccColumnRegex    = regexp.MustCompile(`(?i)(card_number|credit_card|cc_num)`)

	// Data Regex: Detects likely PII inside data strings.
	emailDataRegex = regexp.MustCompile(`(?i)[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}`)
	ssnDataRegex   = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
)

// IsSuspiciousColumn checks if a column name implies PII.
func IsSuspiciousColumn(name string) bool {
	return ssnColumnRegex.MatchString(name) ||
		emailColumnRegex.MatchString(name) ||
		phoneColumnRegex.MatchString(name) ||
		ccColumnRegex.MatchString(name)
}

// ContainsEmail checks if a string contains an email address.
func ContainsEmail(data string) bool {
	return emailDataRegex.MatchString(data)
}

// ContainsSSN checks if a string contains a formatted SSN.
func ContainsSSN(data string) bool {
	return ssnDataRegex.MatchString(data)
}
