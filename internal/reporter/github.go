package reporter

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v62/github"
)

// PostGitHubPRComment posts the failing violations as a Markdown comment to the current PR.
func PostGitHubPRComment(ctx context.Context, receipt *Receipt) error {
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY")
	prStr := os.Getenv("GITHUB_REF_NAME") // Usually "123/merge" in Actions

	if token == "" || repo == "" {
		return nil // Graceful exit if not in GitHub Actions
	}

	// Simple heuristic to extract PR number from ref "refs/pull/123/merge" or "123/merge"
	parts := strings.Split(prStr, "/")
	if len(parts) == 0 {
		return fmt.Errorf("unable to parse PR number from GITHUB_REF_NAME: %s", prStr)
	}

	// In GH actions, GITHUB_REF is refs/pull/:prNumber/merge
	// If ref name is provided, let's assume the first numeric chunk is the PR id
	var prNumber int
	for _, p := range parts {
		if id, err := strconv.Atoi(p); err == nil {
			prNumber = id
			break
		}
	}

	if prNumber == 0 {
		return fmt.Errorf("could not find PR number")
	}

	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 {
		return fmt.Errorf("invalid GITHUB_REPOSITORY format")
	}
	owner := repoParts[0]
	repoName := repoParts[1]

	client := github.NewClient(nil).WithAuthToken(token)
	if apiURL := os.Getenv("GITHUB_API_URL"); apiURL != "" {
		if !strings.HasSuffix(apiURL, "/") {
			apiURL += "/"
		}
		client.BaseURL, _ = url.Parse(apiURL)
	}

	body := "🛑 **Hornfels blocked this migration.**\n\nFound unclassified columns:\n"
	for _, v := range receipt.Violations {
		body += fmt.Sprintf("- `%s.%s`: %s\n  **Run this SQL to fix:**\n  ```sql\n  %s\n  ```\n", v.Table, v.Column, v.Reason, v.ProposedFix)
	}

	comment := &github.IssueComment{Body: &body}
	_, _, err := client.Issues.CreateComment(ctx, owner, repoName, prNumber, comment)
	return err
}
