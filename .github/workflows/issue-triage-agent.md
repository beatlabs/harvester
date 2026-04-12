---
timeout-minutes: 5
strict: true
on:
  issues:
    types: [opened, reopened]
  workflow_dispatch:
permissions:
  issues: read
tools:
  github:
    lockdown: true
    toolsets: [issues, labels]
safe-outputs:
  add-labels:
    max: 4
  add-comment:
    title-prefix: "[triage]"
    max: 1
  update-issue:
    max: 1
imports:
  - shared/mood.md
  - shared/issue-triage-go.md
---

# Issue Triage Agent

List open issues in ${{ github.repository }} that have no labels. For each unlabeled issue, analyze the title and body, then add the most appropriate label from: `bug`, `feature`, `enhancement`, `documentation`, `question`, `help-wanted`, `good-first-issue`, or `config`.

Harvester is a Go HTTP API configuration library. Prioritize labels that reflect configuration sources, seeding, monitoring, retrieval API behavior, usability of the config API, and general library behavior.

Skip issues that:
- Already have any of these labels
- Have been assigned to any user (especially non-bot users)

After adding a label to an issue, mention the issue author in a comment using the shared issue-triage guidance.

**Comment Template**:

```md
Hi @<author> — thanks for opening this issue.

I’ve applied the `<label>` label based on the current description.

If you think a different label would be a better fit, feel free to say so and a maintainer can adjust it.
```

## Labels

- `bug`: Indicates a problem or error in the code that needs fixing.
- `feature`: Represents a new feature request or enhancement to existing functionality.
- `enhancement`: Suggests improvements to existing features or code.
- `documentation`: Pertains to issues related to documentation, such as missing or unclear docs.
- `question`: Used for issues that are asking for clarification or have questions about the project.
- `help-wanted`: Indicates that the issue is a good candidate for external contributions and help.
- `good-first-issue`: Marks issues that are suitable for newcomers to the project, often with simpler scope.
- `config`: Relates to configuration sources, seeding, monitoring, or the config retrieval API.

## Triage Guidance

- Use `bug` for broken behavior, regressions, runtime errors, invalid results, or behavior that contradicts documented expectations.
- Use `feature` for clearly new capabilities or new supported behaviors not currently present in harvester.
- Use `enhancement` for improvements to existing APIs, ergonomics, performance, observability, validation, or developer experience.
- Use `documentation` for missing, outdated, or unclear README, examples, API docs, or usage guidance.
- Use `question` for support-style requests, clarifications, or issues that do not yet describe a concrete defect or change request.
- Use `help-wanted` only when the issue is clearly actionable and suitable for outside contribution.
- Use `good-first-issue` only when the change appears well-scoped, low-risk, and approachable for a first-time contributor.
- Use `config` when the core topic is about configuration loading, value retrieval, provider behavior, HTTP/API-based config access, seeding, or monitoring.

Choose the single best primary label when possible. Only add multiple labels when they provide distinct value and remain within the safe-output limits.

## Batch Optimization

- Work in batches to reduce GitHub API calls.
- Start by identifying open issues with no labels and no assignees.
- Triage only the qualifying issues from that batch.
- Apply labels and post at most one triage comment per issue.
- Avoid unnecessary updates when the issue content is ambiguous; prefer `question` in borderline cases.
