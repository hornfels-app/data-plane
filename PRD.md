# Hornfels - Product Requirements Document (PRD)

## 1. Product Identity
**Name:** Hornfels
**Tagline:** The Developer's PII Control Layer for Database Compliance.
**Philosophy:** 90% thinking / 10% speaking. Brutally realistic. Ephemeral, stateless, and air-gapped.

## 2. Problem Statement
Developers hate security tools because they are hard to set up, produce false positives, and block PRs without providing a clear fix. Startups need SOC2 compliance (knowing where their PII is) but refuse to deploy heavy Python ML pipelines or expensive SaaS proxies on day one.

## 3. Target Audience
- Fast-moving startups using Go, Node.js (TypeScript), or Rust.
- Primary Databases: PostgreSQL and MySQL.
- Primary ORMs: Prisma (Node.js) and raw SQL migrations (Flyway/Liquibase).
- Primary Environments: GitHub Actions / CI Pipelines.

## 4. Phase 1 Scope (Data Plane Only)
- A single, static Go binary that runs in CI.
- Connects to an ephemeral CI database or parses Prisma schemas.
- Scans schemas for unclassified columns.
- Blocks Pull Requests if a developer adds a column without explicitly tagging it as PII or non-PII.
- Zero network calls to outside APIs (completely air-gapped).

## 5. Core Requirements
### 5.1. Baseline Generation
- **Command:** `hornfels baseline`
- **Behavior:** Scans the database and writes `.hornfels-baseline.yaml`.
- **Purpose:** Ignores all existing legacy columns so adoption takes exactly 30 seconds. Zero Day-1 false positives.

### 5.2. Strict Schema Mode (Default)
- **Command:** `hornfels check`
- **Behavior:** Queries the live database `information_schema` (or `pg_description`).
- **Policy:** Every column added *after* the baseline must have a SQL comment containing `[hornfels: pii=true]` or `[hornfels: pii=false]`.

### 5.3. Prisma Native Support
- **Command:** `hornfels check --prisma`
- **Behavior:** Parses `schema.prisma` natively, looking for `/// [hornfels: pii=true]` comments without requiring a live database.

### 5.4. Data Scan Mode (Bonus Layer)
- **Command:** `hornfels check --scan-data`
- **Behavior:** Pulls exactly `LIMIT 100` rows from the CI database. Runs native Go heuristics (Regex and Luhn algorithms) to detect SSNs, Credit Cards, and Emails in JSONB/unstructured columns.
- **Policy:** Never sends data out of the CI environment. Memory is garbage collected immediately.

### 5.5. AI and Developer Experience (DevEx)
- **PR Comments:** Uses `GITHUB_TOKEN` to post the exact, copy-pasteable SQL/Prisma fix in the Pull Request.
- **AI Init:** `hornfels init` generates a `.cursorrules` / `CLAUDE.md` snippet so local AI agents automatically fix Hornfels errors.

## 6. Phase 2 Transition (The Wedge)
- **Feature:** `hornfels login` and `hornfels sync`
- **Goal:** Upsell the startup when they start their SOC2 audit. We sync the local `hornfels-receipt.json` to our SaaS API and map it to Vanta/Drata automatically.
