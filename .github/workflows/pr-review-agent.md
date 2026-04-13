---
description: On-demand deep PR review triggered by /review slash command.
on:
  slash_command:
    command: /review
tools: github[pull_requests, repos], bash
imports:
  - shared/mood.md
  - shared/go-ci.md
  - shared/go-security.md
engine: copilot
strict: true
timeout-minutes: 15
permissions:
  contents: read
  pull-requests: read
safe-outputs:
  create-pull-request-review-comment:
    max: 15
    title-prefix: "[review]"
  add-comment:
    title-prefix: "[review]"
    max: 1
  add-labels:
    max: 2
---

## Scope

Review the current pull request diff comprehensively.

Fetch the full diff for the current PR and analyze every changed file. Cover all changed Go code, tests, configuration, and any other touched files that may affect behavior, safety, or compatibility.

## Review categories

### Correctness

Look for logic errors, missing edge cases, off-by-one mistakes, nil or zero-value mishandling, goroutine leaks, broken invariants, and context cancellation gaps.

### Security

Use `shared/go-security.md` guidance. Look for SQL injection, path traversal, hardcoded secrets, insecure crypto, auth or authorization gaps, and missing input validation.

### Performance

Look for unnecessary allocations, O(n²) patterns in hot paths, missing connection pooling, unbounded slice or map growth, excessive locking, and avoidable blocking work.

### Testing gaps

Look for changed logic without matching test updates, untested error paths, missing table-driven coverage, and missing regression tests for risky branches.

### API design

Look for breaking changes to exported symbols, inconsistent naming, unclear contracts, and missing godoc on exported types or functions introduced by the PR.

## Boundaries

This review complements `code-simplifier.md`.

- Code-simplifier covers simplification, dead code, idiomatic Go, and boolean expression cleanup.
- This review must focus only on correctness, security, performance, testing gaps, and API design.
- Do not comment on simplification opportunities.
- Do not suggest dead code cleanup unless it causes a correctness, security, or API issue.
- Do not suggest import reorganization or style formatting.
- Avoid subjective feedback.

## Commenting rules

- Post line-level review comments for specific findings only.
- Prefix every line-level review comment with `[review]`.
- Keep each comment concise, actionable, and tied to one issue.
- Do not duplicate the same issue across multiple comments.
- Prioritize high-confidence findings.

## Summary comment

Post exactly one final summary comment.

The summary must include:

- overall assessment: `approve`, `request-changes`, or `comment`
- count of findings by category
- top 3 highest-priority items if there are more than 5 findings

If the PR is clean across all review categories, post one summary comment stating that the review found no issues and do not add labels.

## Labels

- Add `security-concern` if any security finding is present.
- Add `needs-tests` if any testing gap is present.
- Do not add labels otherwise.

## Execution rules

- Review the actual PR diff, not the base branch broadly.
- Analyze every changed file before deciding on the summary assessment.
- Keep feedback broad and summary-oriented while staying specific on concrete defects.
- Prefer fewer, higher-signal comments over exhaustive low-value notes.
