---
# Go CI Shared Import
# Standardized Go CI execution shape for Taskfile-based repos
#
# Usage:
#   imports:
#     - shared/go-ci.md
#
# This import provides:
# - Go version detection from go.mod
# - Standardized Taskfile task names for lint, test, and CI
# - Network allowlist for Go module proxy and checksum database
#
# Note: This import does NOT define lint rules. Each repo maintains
# its own .golangci.yml with repo-specific linter configuration.
# Patron uses 50+ linters; harvester uses 39 linters.

network:
  allowed:
    - defaults
    - proxy.golang.org
    - sum.golang.org
    - pkg.go.dev

tools:
  bash:
    - "go *"
    - "task *"
    - "git"
---

# Go CI Environment

This repo uses [Task](https://taskfile.dev/) as the build system. All CI operations are invoked through Taskfile targets.

## Go Version Detection

```bash
# Detect Go version from go.mod
go mod edit -json | jq -r .Go
```

## Standard Task Targets

| Target | Purpose |
|--------|---------|
| `task lint` | Run golangci-lint with repo-local config |
| `task test` | Run unit tests with race detector |
| `task testint` | Run integration tests |
| `task ci` | Full CI pipeline (lint + test) |
| `task fmt` | Format Go source files |
| `task deps-start` | Start Docker dependencies for integration tests |
| `task deps-stop` | Stop Docker dependencies |

## Lint Configuration

Each repository maintains its own `.golangci.yml`. This shared import does not override or define lint rules. Reference the repo-local config for the active linter set.

## Network Access

The following domains are allowed for Go module operations:

- `proxy.golang.org` — Go module proxy
- `sum.golang.org` — Go checksum database
- `pkg.go.dev` — Go package documentation (read-only reference)
