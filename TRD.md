# Hornfels - Technical Requirements Document (TRD)

## 1. System Architecture
Hornfels Phase 1 is a stateless Go binary designed for CI/CD execution.

**Inputs:**
- `DATABASE_URL` (Environment Variable)
- `GITHUB_TOKEN` (Environment Variable, optional for PR comments)
- `.hornfels.yaml` (Configuration)
- `.hornfels-baseline.yaml` (Baseline state)

**Outputs:**
- Exit Code: `0` (Pass) or `1` (Fail)
- Local Artifact: `hornfels-receipt.json`
- Network: GitHub API POST (if PR fails and token is present)

## 2. Package Structure (Go 1.21+)
```text
github.com/hornfels/hornfels
├── cmd/
│   └── hornfels/          # Cobra CLI commands (check, baseline, init, login)
├── internal/
│   ├── scanner/           # DB Abstraction Layer
│   │   ├── postgres.go    # Uses pgx/v5 to query pg_class / pg_description
│   │   ├── mysql.go       # Uses go-sql-driver/mysql to query information_schema
│   │   └── prisma.go      # Regex/AST parser for schema.prisma
│   ├── policy/            # Core logic engine
│   │   ├── yaml.go        # Parses .hornfels.yaml
│   │   └── baseline.go    # Merges current schema with baseline ignores
│   ├── heuristics/        # Data scanning math
│   │   ├── luhn.go        # Credit card checksum validation
│   │   └── regex.go       # Compiled regexes for SSN, Email, Phone
│   ├── reporter/          # Output formatters
│   │   ├── stdout.go      # Terminal UI (pterm or lipgloss)
│   │   ├── json.go        # Generates hornfels-receipt.json
│   │   └── github.go      # Uses google/go-github to post PR comments
│   └── auth/              # Phase 2 stub for SaaS API tokens
├── pkg/                   # Publicly exportable utilities
└── go.mod
```

## 3. Database Queries
### 3.1 PostgreSQL Catalog Query
To find unclassified columns, the `scanner` package executes:
```sql
SELECT 
    c.relname AS table_name,
    a.attname AS column_name,
    d.description AS pii_tag,
    t.typname AS data_type
FROM pg_class c
JOIN pg_attribute a ON c.oid = a.attrelid
JOIN pg_type t ON a.atttypid = t.oid
LEFT JOIN pg_description d ON c.oid = d.objoid AND a.attnum = d.objsubid
WHERE c.relkind = 'r' AND a.attnum > 0 AND c.relnamespace = 'public'::regnamespace;
```

## 4. Heuristics & Data Scanning
When `--scan-data` is passed, the engine runs `SELECT * FROM <table> LIMIT 100`.
- **Constraint:** Native Go logic only. No Python, no CGO, no SpaCy, no heavy NLP.
- **Algorithms:**
  - Luhn Algorithm (Mod 10) for Credit Cards.
  - Standard RFC 5322 regex for Emails.
  - Regional regex for SSNs.

## 5. Security & Constraints
- **Air-Gapped:** The binary must not contain any network egress code other than the explicitly configured GitHub API endpoint.
- **Stateless:** Must not persist data locally other than the `hornfels-receipt.json` artifact.
- **Dependency Minimization:** Avoid heavy frameworks. Use standard library where possible. Use `jackc/pgx` for Postgres.
