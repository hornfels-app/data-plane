package main_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEndToEnd(t *testing.T) {
	ctx := context.Background()

	// 1. Compile the Hornfels binary
	binPath := filepath.Join(t.TempDir(), "hornfels")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/hornfels")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build hornfels: %v\n%s", err, out)
	}

	// 2. Set up GitHub Mock API
	ghMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/repos/testowner/testrepo/issues/123/comments") {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 1}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ghMock.Close()

	// 3. Set up Postgres via Testcontainers
	dbName := "hornfels_test"
	dbUser := "devuser"
	dbPassword := "devpassword"
	
	initScript := filepath.Join(t.TempDir(), "init.sql")
	err := os.WriteFile(initScript, []byte(`
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50),
    email VARCHAR(255),
    ssn VARCHAR(20)
);
COMMENT ON COLUMN users.email IS '[hornfels: pii=true] User email address';
COMMENT ON COLUMN users.ssn IS '[hornfels: pii=false] Not an SSN';
INSERT INTO users (first_name, email, ssn) VALUES ('Alice', 'alice@example.com', '123-45-6789');
	`), 0644)
	if err != nil {
		t.Fatalf("Failed to write init.sql: %v", err)
	}

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithInitScripts(initScript),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(20*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	// 4. Helper to run CLI
	runCLI := func(args ...string) (string, error) {
		cmd := exec.Command(binPath, args...)
		cmd.Env = append(os.Environ(),
			"DATABASE_URL="+connStr,
			"GITHUB_TOKEN=fake_token",
			"GITHUB_REPOSITORY=testowner/testrepo",
			"GITHUB_REF_NAME=123/merge",
			"GITHUB_API_URL="+ghMock.URL,
		)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	// 5. Test Step: Init
	t.Log("Testing 'hornfels init'...")
	if out, err := runCLI("init"); err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}

	// 6. Test Step: Baseline
	t.Log("Testing 'hornfels baseline'...")
	if out, err := runCLI("baseline"); err != nil {
		t.Fatalf("baseline failed: %v\n%s", err, out)
	}

	// 7. Test Step: Check (Should Pass because everything is baselined)
	t.Log("Testing 'hornfels check'...")
	out, err := runCLI("check")
	if err != nil {
		t.Fatalf("check failed when it should have passed (due to baseline): %v\n%s", err, out)
	}
	if !strings.Contains(out, "Hornfels Check Passed") {
		t.Errorf("Expected PASS output, got: %s", out)
	}

	// 8. Delete Baseline to test Strict Mode
	os.Remove(".hornfels-baseline.yaml")

	// 9. Test Step: Check --scan-data (Should Fail because SSN is tagged false but contains data)
	t.Log("Testing 'hornfels check --scan-data'...")
	out, err = runCLI("check", "--scan-data")
	if err == nil {
		t.Fatalf("check --scan-data should have failed, but it passed.\n%s", out)
	}
	if !strings.Contains(out, "Hornfels Check Failed") {
		t.Errorf("Expected FAIL output, got: %s", out)
	}
	if !strings.Contains(out, "Column tagged as pii=false but sampled data contains SSN") {
		t.Errorf("Did not find heuristic failure reason in output: %s", out)
	}
	
	// Cleanup config files
	os.Remove(".hornfels.yaml")
	os.Remove(".cursorrules")
	os.Remove("hornfels-receipt.json")
}
