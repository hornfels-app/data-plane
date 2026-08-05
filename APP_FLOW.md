# Hornfels - App Flow & User Journey

## Flow 1: Day-1 Onboarding (The Setup)
1. **Installation:** Developer downloads the binary or runs a homebrew/curl install script.
2. **Initialization:** Developer runs `hornfels init`.
   - Tool creates `.hornfels.yaml` with default settings.
   - Tool injects a `.cursorrules` file into the project root to enable AI agents.
3. **Baselining:** Developer runs `hornfels baseline`.
   - Hornfels connects to their local or staging database.
   - It sees 500 existing tables and 4,000 columns with no PII tags.
   - It generates `.hornfels-baseline.yaml`, ignoring all 4,000 columns.
4. **Commit:** Developer commits `.hornfels.yaml` and `.hornfels-baseline.yaml` to Git and merges to `main`.
   - *Result: Setup is complete. Zero disruption to existing workflows.*

## Flow 2: The Blocked Pull Request (Strict Mode)
1. **Action:** A junior developer creates a new feature branch and adds a column `user_medical_history` to the database. They push the branch.
2. **CI Execution:** GitHub Actions runs `hornfels check`.
3. **Detection:** Hornfels compares the live CI schema against the baseline. It finds `user_medical_history` is new and has no `[hornfels: pii=X]` comment.
4. **Enforcement:**
   - Hornfels exits with status `1` (failing the CI build).
   - Hornfels uses the CI `GITHUB_TOKEN` to post a comment directly on the PR:
     > 🛑 **Hornfels blocked this migration.**
     > Column `user_medical_history` lacks a PII classification. 
     > **Run this SQL to fix:**
     > `COMMENT ON COLUMN users.user_medical_history IS '[hornfels: pii=true]';`
5. **Resolution:** The developer copies the SQL, applies it to their migration file, pushes, and the CI passes.

## Flow 3: The AI Auto-Fix (Cursor/Claude Code)
1. **Action:** A developer adds a new table using Prisma and pushes. The PR fails because of Hornfels.
2. **AI Intervention:** The developer opens Cursor/Claude Code and types: *"Fix the Hornfels CI error."*
3. **Resolution:** Because of the `.cursorrules` injected during `hornfels init`, the AI knows exactly how Hornfels works. The AI immediately adds `/// [hornfels: pii=false]` to the `schema.prisma` file and commits the code. The developer did exactly zero work.

## Flow 4: The Phase 2 Upgrade (Monetization)
1. **Trigger:** Six months later, the CTO is preparing for a SOC2 audit and needs proof of PII tracking.
2. **Action:** The CTO goes to `hornfels.app`, buys a license, and gets an API key.
3. **Execution:** The CTO runs `hornfels login` locally and adds the `HORNFELS_API_KEY` to GitHub Actions.
4. **Resolution:** Every time `hornfels check` runs in CI, it now syncs the cryptographically signed `hornfels-receipt.json` to the Hornfels SaaS platform. The Hornfels SaaS platform pushes these receipts directly into Vanta/Drata via their APIs, providing continuous compliance evidence.
