---
description: Review Go pull request diffs and suggest high-confidence simplifications.
on:
  pull_request:
    types: [opened, synchronize]
    paths: ["**.go"]
tools:
  github:
    toolsets: [pull_requests, repos]
  bash:
    - "*"
imports:
  - shared/mood.md
  - shared/go-ci.md
engine: copilot
strict: true
timeout-minutes: 10
permissions:
  contents: read
  pull-requests: read
safe-outputs:
  create-pull-request-review-comment:
    max: 10
  add-labels:
    max: 1
---

Review the current pull request's Go diff only.

Analyze the changed Go lines and identify concrete, high-confidence simplification opportunities. Focus on:
- unnecessary complexity such as overly nested if/else blocks or redundant nil checks
- idiomatic Go improvements, including better error-handling structure and range-over-func patterns when appropriate
- dead code or unreachable branches
- unnecessary type conversions
- simplifiable boolean expressions

Rules:
- Post review comments on specific lines only, using the `[simplify]` prefix.
- Keep comments concise: one suggestion per comment, with a small code snippet showing the simpler version.
- Respect the repository's local `.golangci.yml`. Do not suggest anything that conflicts with the local lint configuration.
- Prioritize only high-confidence suggestions. Avoid subjective style preferences.
- Keep the total number of comments at 10 or fewer. Prioritize the highest-impact suggestions.
- Do not suggest changes to test files unless the simplification is about test structure rather than assertions.
- Do not suggest dependency changes.
- Do not suggest import reorganization.

If you make 3 or more actionable suggestions, add the `simplification` label once.

If there are no clear, high-confidence simplifications, do not comment.
