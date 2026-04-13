---
description: Generate tag-based GitHub Release notes for harvester.
on:
  push:
    tags:
      - "v*"
tools:
  - github[repos, pull_requests, issues]
  - bash
imports:
  - shared/mood.md
  - shared/release-notes-template.md
engine: copilot
strict: true
timeout-minutes: 15
permissions:
  contents: read
  pull-requests: read
  issues: read
safe-outputs:
  update-release:
    max: 1
  add-comment:
    max: 1
---

# Release Notes Generator

Generate release notes for the pushed tag and update the matching GitHub Release body.

## Requirements

1. Detect the pushed tag from the workflow context.
2. Find the previous tag. If there is no previous tag, treat this as the first release.
3. Collect all commits between the previous tag and the pushed tag. If this is the first release, use all commits reachable by the pushed tag.
4. Map commits to merged pull requests where possible using commit metadata, associated PRs, and issue references.
5. Categorize changes using pull request labels first, then commit messages when labels are unavailable:
   - `breaking-change`, `api-change` → `Breaking Changes`
   - `feature` → `New Features`
   - `enhancement` → `Improvements`
   - `bug` → `Bug Fixes`
   - `dependencies`, `dependabot` → `Dependencies`
6. Build the release body from the shared `release-notes-template.md` import.
7. Keep the `Breaking Changes` section at the top and include migration guidance for every breaking change.
8. In the `Dependencies` section, summarize `go.mod` and `go.sum` changes when present.
9. In the `Contributors` section, list first-time contributors by checking whether each commit author has prior commits in the repository before this release range.
10. Handle edge cases:
    - no previous tag
    - no merged PRs for some or all commits
    - commit-only releases where PR mapping is unavailable

## Execution Notes

- Use `bash` for tag diffing, commit range calculation, and assembling the final release body.
- Use GitHub repository, pull request, and issue data to enrich entries with PR numbers, links, labels, and short descriptions.
- For commit-only entries with no PR mapping, still include them in the most appropriate section using the commit message and commit link.
- Keep entries concise and link to the PR when available.
- Do not assume any release tooling beyond tags and GitHub Releases.

## Outputs

- Use `update-release` once to populate the GitHub Release body for the pushed tag.
- If the GitHub Release does not exist yet, use `add-comment` once to note that the release was not found and could not be updated.
