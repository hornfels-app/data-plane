package heuristics

import (
	"strconv"
	"strings"
)

// IsValidLuhn checks if a given credit card string passes the Luhn algorithm.
// It ignores spaces and dashes.
func IsValidLuhn(cc string) bool {
	cc = strings.ReplaceAll(cc, " ", "")
	cc = strings.ReplaceAll(cc, "-", "")

	if len(cc) < 13 || len(cc) > 19 {
		return false
	}

	sum := 0
	double := false

	// Iterate backwards
	for i := len(cc) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(cc[i]))
		if err != nil {
			return false // Not a digit
		}

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
