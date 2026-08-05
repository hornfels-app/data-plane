package heuristics

import (
	"testing"
)

func TestRegex(t *testing.T) {
	if !IsSuspiciousColumn("user_ssn") {
		t.Errorf("Expected user_ssn to be suspicious")
	}
	if IsSuspiciousColumn("created_at") {
		t.Errorf("Expected created_at NOT to be suspicious")
	}

	if !ContainsEmail("contact test@example.com for more") {
		t.Errorf("Expected to find email")
	}
	if !ContainsSSN("My SSN is 123-45-6789") {
		t.Errorf("Expected to find SSN")
	}
}

func TestLuhn(t *testing.T) {
	// 79927398713 is a valid Luhn (11 digits, but let's test a standard 16 digit one)
	// Example valid visa: 4000 0000 0000 0000 (wait, 4000000000000000 doesn't pass Luhn if 4 is at pos 15, let's just use 4111 1111 1111 1111)
	if !IsValidLuhn("4111 1111 1111 1111") {
		t.Errorf("Expected 4111 1111 1111 1111 to pass Luhn")
	}

	if IsValidLuhn("4111 1111 1111 1112") {
		t.Errorf("Expected 4111 1111 1111 1112 to fail Luhn")
	}
}
