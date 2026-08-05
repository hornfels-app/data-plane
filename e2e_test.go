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
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func buildCLI(t *testing.T) string {
	binPath := filepath.Join(t.TempDir(), "hornfels")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/hornfels")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build hornfels: %v\n%s", err, out)
	}
	return binPath
}

func setupMockGitHub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/repos/testowner/testrepo/issues/123/comments") {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 1}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func runCLI(binPath, connStr, ghMockURL string, args ...string) (string, error) {
	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(),
		"DATABASE_URL="+connStr,
		"GITHUB_TOKEN=fake_token",
		"GITHUB_REPOSITORY=testowner/testrepo",
		"GITHUB_REF_NAME=123/merge",
		"GITHUB_API_URL="+ghMockURL,
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestPostgresEndToEnd(t *testing.T) {
	ctx := context.Background()
	binPath := buildCLI(t)
	ghMock := setupMockGitHub()
	defer ghMock.Close()

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
		postgres.WithDatabase("hornfels_test"),
		postgres.WithUsername("devuser"),
		postgres.WithPassword("devpassword"),
		postgres.WithInitScripts(initScript),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(20*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer postgresContainer.Terminate(ctx)

	connStr, _ := postgresContainer.ConnectionString(ctx, "sslmode=disable")

	// 1. Init
	if _, err := runCLI(binPath, connStr, ghMock.URL, "init"); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// 2. Baseline
	if _, err := runCLI(binPath, connStr, ghMock.URL, "baseline"); err != nil {
		t.Fatalf("baseline failed: %v", err)
	}

	// 3. Check (Pass)
	out, err := runCLI(binPath, connStr, ghMock.URL, "check")
	if err != nil || !strings.Contains(out, "Hornfels Check Passed") {
		t.Fatalf("check failed when it should pass: %v\n%s", err, out)
	}

	// 4. Strict Mode
	os.Remove(".hornfels-baseline.yaml")
	out, err = runCLI(binPath, connStr, ghMock.URL, "check", "--scan-data")
	if err == nil || !strings.Contains(out, "Hornfels Check Failed") {
		t.Fatalf("check --scan-data should have failed.\n%s", out)
	}

	os.Remove(".hornfels.yaml")
	os.Remove(".cursorrules")
}

func TestMySQLEndToEnd(t *testing.T) {
	ctx := context.Background()
	binPath := buildCLI(t)
	ghMock := setupMockGitHub()
	defer ghMock.Close()

	initScript := filepath.Join(t.TempDir(), "init.sql")
	err := os.WriteFile(initScript, []byte(`
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50),
    email VARCHAR(255),
    ssn VARCHAR(20)
);
ALTER TABLE users MODIFY COLUMN email VARCHAR(255) COMMENT '[hornfels: pii=true] User email address';
ALTER TABLE users MODIFY COLUMN ssn VARCHAR(20) COMMENT '[hornfels: pii=false] Not an SSN';
INSERT INTO users (first_name, email, ssn) VALUES ('Alice', 'alice@example.com', '123-45-6789');
	`), 0644)
	if err != nil {
		t.Fatalf("Failed to write init.sql: %v", err)
	}

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("hornfels_test"),
		mysql.WithUsername("devuser"),
		mysql.WithPassword("devpassword"),
		mysql.WithScripts(initScript),
	)
	if err != nil {
		t.Fatalf("failed to start mysql container: %s", err)
	}
	defer mysqlContainer.Terminate(ctx)

	connStr, _ := mysqlContainer.ConnectionString(ctx)

	// 1. Init
	if _, err := runCLI(binPath, connStr, ghMock.URL, "init"); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// 2. Baseline
	if _, err := runCLI(binPath, connStr, ghMock.URL, "baseline"); err != nil {
		t.Fatalf("baseline failed: %v", err)
	}

	// 3. Check (Pass)
	out, err := runCLI(binPath, connStr, ghMock.URL, "check")
	if err != nil || !strings.Contains(out, "Hornfels Check Passed") {
		t.Fatalf("check failed when it should pass: %v\n%s", err, out)
	}

	// 4. Strict Mode
	os.Remove(".hornfels-baseline.yaml")
	out, err = runCLI(binPath, connStr, ghMock.URL, "check", "--scan-data")
	if err == nil || !strings.Contains(out, "Hornfels Check Failed") {
		t.Fatalf("check --scan-data should have failed.\n%s", out)
	}

	os.Remove(".hornfels.yaml")
	os.Remove(".cursorrules")
}

func TestPrismaEndToEnd(t *testing.T) {
	binPath := buildCLI(t)
	ghMock := setupMockGitHub()
	defer ghMock.Close()

	os.WriteFile("schema.prisma", []byte(`
model User {
  id        Int      @id @default(autoincrement())
  email     String   /// [hornfels: pii=true]
  ssn       String   // Missing the hornfels tag!
}
	`), 0644)

	runCLI(binPath, "", ghMock.URL, "init")

	out, err := runCLI(binPath, "", ghMock.URL, "check", "--prisma")
	if err == nil || !strings.Contains(out, "Hornfels Check Failed") {
		t.Fatalf("Prisma check should have failed on untagged SSN.\n%s", out)
	}

	os.Remove("schema.prisma")
	os.Remove(".hornfels.yaml")
	os.Remove(".cursorrules")
}
