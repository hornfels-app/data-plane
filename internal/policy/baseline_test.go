package policy

import (
	"testing"
)

func TestBaselineIsIgnored(t *testing.T) {
	b := &Baseline{
		IgnoredColumns: map[string][]string{
			"users": {"id", "created_at"},
		},
	}

	if !b.IsIgnored("users", "id") {
		t.Errorf("Expected users.id to be ignored")
	}

	if b.IsIgnored("users", "email") {
		t.Errorf("Expected users.email NOT to be ignored")
	}

	if b.IsIgnored("posts", "id") {
		t.Errorf("Expected posts.id NOT to be ignored")
	}
}
