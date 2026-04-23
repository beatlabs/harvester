---
# Go Issue Triage Shared Import
# Label taxonomy and classification logic for Go repositories
#
# Usage:
#   imports:
#     - shared/issue-triage-go.md
#
# This import provides:
# - Standardized label taxonomy for Go repos
# - Go-specific issue pattern matching rules
# - Duplicate detection heuristics
# - Staleness and auto-close criteria
#
# Repo context:
# - Patron already has an Issue Triage Agent using this taxonomy
# - Harvester is the primary new deployment target
# - This import ensures consistent labeling across both repos
---

# Go Issue Triage Rules

## Label Taxonomy

### Type Labels
| Label | When to apply |
|-------|--------------|
| `bug` | Confirmed or suspected defect: panics, nil pointers, incorrect behavior |
| `enhancement` | Improvement to existing functionality |
| `feature` | New capability request |
| `question` | Seeking clarification or usage help |
| `documentation` | Missing, incorrect, or unclear docs |
| `performance` | Slowness, excessive memory/CPU, inefficient patterns |

### Area Labels
| Label | When to apply |
|-------|--------------|
| `api` | Exported types, interfaces, or function signatures |
| `core` | Internal implementation, non-exported code |
| `deps` | Dependency versions, go.mod changes, module conflicts |
| `ci` | CI pipeline, GitHub Actions, build system |
| `test` | Test failures, coverage gaps, test infrastructure |

### Priority Labels
| Label | Criteria |
|-------|----------|
| `P0-critical` | Data loss, security vulnerability, complete breakage |
| `P1-high` | Panic/nil pointer, broken core functionality |
| `P2-medium` | Degraded functionality, workaround exists |
| `P3-low` | Cosmetic, minor inconvenience, nice-to-have |

### Status Labels
| Label | Purpose |
|-------|---------|
| `needs-info` | Issue lacks reproduction steps or environment details |
| `good-first-issue` | Clear scope, well-defined, suitable for newcomers |
| `help-wanted` | Maintainer cannot prioritize, community contribution welcome |
| `wontfix` | Intentional behavior or out of scope |

## Go-Specific Pattern Matching

Apply these rules based on issue content:

| Signal in issue | Type | Priority |
|----------------|------|----------|
| Panic, nil pointer, segfault | `bug` | `P1-high` |
| `go.mod`, `go.sum`, dependency version | `deps` | `P2-medium` |
| Exported type/interface change | `api` | Flag for breaking change review |
| Test failure, coverage drop | `test` | `P2-medium` |
| Slow, timeout, memory leak | `performance` | `P2-medium` |
| "How do I", "is it possible" | `question` | `P3-low` |

## Duplicate Detection

Before labeling, check for similar existing issues:
1. Compare title against the last 30 open issues using keyword overlap
2. If >70% keyword match, add a comment linking to the potential duplicate
3. Do NOT auto-close — let the maintainer decide

## Staleness Criteria

| Condition | Action |
|-----------|--------|
| >90 days inactive + `P3-low` | Add `stale` label, comment warning |
| >180 days inactive + `P3-low` + no activity after warning | Close with explanation |
| Has `keep-open` label | Exempt from staleness rules |
