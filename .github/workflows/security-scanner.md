---
on:
  schedule:
    - cron: "0 4 * * 1"
  workflow_dispatch:
  push:
    branches: [master]
    paths: ["go.mod", "go.sum"]
permissions:
  contents: read
  issues: read
tools:
  github: [issues, repos, dependabot]
  bash: true
safe-outputs:
  create-issue:
    title-prefix: "[security]"
    labels: ["security", "P1-high"]
    max: 3
    close-older-issues: true
    skip-if-match: "\\[security\\].*CVE-.*open"
  add-labels:
    max: 2
imports:
  - shared/mood.md
  - shared/go-security.md
engine: copilot
strict: true
timeout-minutes: 20
network:
  allowed: [defaults, vuln.go.dev]
---

# Security Scanner

Run a focused Go security scan for this repository.

## Objectives

- Run `govulncheck ./...` to detect known vulnerabilities in dependencies.
- Run `gosec ./...` for static security analysis of Go code.
- Audit `go.mod` and `go.sum` for dependency health signals.

## Severity classification

- **Critical**: Known exploitable CVE in a direct dependency.
- **High**: CVE in a direct dependency that is not yet exploitable in this codebase.
- **Medium**: CVE in a transitive dependency.
- **Low / Informational**: `gosec` style findings and best-practice warnings.

## Issue creation rules

- Create issues only for **Critical** and **High** findings.
- Deduplicate by CVE; do not create duplicate issues for the same vulnerability.
- Include the **CVE ID in the issue title** for deduplication via `skip-if-match`.
- Create at most **3 issues per run**, prioritized by severity.
- Add the `security` label to all findings.
- Add the `dependency` label when the finding is dependency-related.
- Clearly separate actionable findings from informational ones in the issue body.

## Required issue content

For each created issue, include:

- CVE ID in the title.
- Affected dependency and version.
- Whether the dependency is direct or transitive.
- Recommended remediation: upgrade, replace, or mitigate.

## Repository context

If the repository already has `gosec` coverage through its lint configuration (as patron does), treat this workflow as complementary rather than duplicative.
