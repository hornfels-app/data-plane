package scanner

import (
	"testing"
)

func TestHasHornfelsTag(t *testing.T) {
	tests := []struct {
		comment string
		hasTag  bool
		isPII   bool
	}{
		{"1=active, 2=inactive [hornfels: pii=false]", true, false},
		{"Just a normal comment", false, false},
		{"[hornfels: pii=true] user email", true, true},
		{"[HORNFELS: PII=TRUE]", true, true}, // case insensitive
		{"[hornfels: pii=false]", true, false},
		{"[hornfels: pii=none]", false, false}, // invalid bool
	}

	for _, tt := range tests {
		hasTag, isPII := HasHornfelsTag(tt.comment)
		if hasTag != tt.hasTag || isPII != tt.isPII {
			t.Errorf("HasHornfelsTag(%q) = (%v, %v), want (%v, %v)", tt.comment, hasTag, isPII, tt.hasTag, tt.isPII)
		}
	}
}
