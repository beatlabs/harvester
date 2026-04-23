---
# Go Security Scanning Shared Import
# Guidance for govulncheck and gosec in Go repositories
#
# Usage:
#   imports:
#     - shared/go-security.md
#
# This import provides:
# - govulncheck invocation and output parsing
# - gosec invocation with JSON output and severity bucketing
# - Dependency audit patterns for go.mod / go.sum changes
# - Network allowlist for vuln.go.dev
#
# Repo baselines:
# - Patron: already gets gosec signal through golangci-lint integration
# - Harvester: should rely on this import for explicit security scanning

network:
  allowed:
    - defaults
    - vuln.go.dev
    - proxy.golang.org
    - sum.golang.org

tools:
  bash:
    - "go *"
    - "govulncheck *"
    - "gosec *"
    - "jq *"
    - "git"
---

# Go Security Scanning

## govulncheck

Scan for known vulnerabilities in Go dependencies and source code.

```bash
# Install
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run against all packages
govulncheck ./...
```

Output includes:
- Vulnerability ID (e.g., GO-2024-XXXX)
- Affected module and version
- Whether the vulnerable symbol is actually called in code

## gosec

Scan Go source for common security issues.

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run with JSON output for structured processing
gosec -fmt json -out /tmp/gh-aw/gosec-results.json ./...
```

## Severity Bucketing

When reporting findings, classify by severity:

| Severity | Action | SLA |
|----------|--------|-----|
| Critical | Immediate issue, label `P0-critical` | 24h |
| High | Issue with `P1-high`, assign owner | 7 days |
| Medium | Issue with `P2-medium` | 30 days |
| Low | Track only, no issue unless pattern | — |

## Dependency Audit

Check `go.mod` and `go.sum` for problematic patterns:

```bash
# List all direct dependencies with versions
go list -m -json all | jq -r 'select(.Indirect != true) | "\(.Path) \(.Version)"'

# Check for replaced modules (potential supply chain concern)
go mod edit -json | jq -r '.Replace[]? | "\(.Old.Path) -> \(.New.Path)"'
```

## Network Access

- `vuln.go.dev` — Go vulnerability database
- `proxy.golang.org` — Module downloads for analysis
- `sum.golang.org` — Checksum verification
