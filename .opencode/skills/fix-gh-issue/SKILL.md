---
name: fix-gh-issue
description: Fix a single GitHub issue end-to-end using wt. Creates a worktree, implements the fix, waits for CI, squash-merges, and removes the worktree. Trigger phrases: 'fix issue', 'fix gh issue', 'work on issue', 'start issue', 'launch issue'.
argument-hint: "<issue-number> [branch-name] [prompt]"
---

# fix-gh-issue Skill

Orchestrate a single GitHub issue through the full lifecycle: worktree → implement → PR → CI → squash-merge → cleanup.

Works in any git repository with a GitHub remote.

## Inputs

Extract from the user message:
- `ISSUE` — GitHub issue number (required, e.g. `42`)
- `BRANCH` — branch name (optional — auto-derived if omitted)
- `PROMPT` — custom prompt describing the fix (optional — default below)

If `ISSUE` is missing, ask before proceeding.

## Derived values

Resolve at runtime — do not hardcode paths or repo names:

```bash
REPO_PATH=$(git rev-parse --show-toplevel)
REPO_NAME=$(basename "$REPO_PATH")
```

If `BRANCH` was not provided, derive it from the issue title:

```bash
BRANCH=$(gh issue view "$ISSUE" --json number,title \
  --jq '"fix/\(.number)-\(.title | gsub("^[a-z]+: "; "") | ascii_downcase | gsub("[^a-z0-9]+"; "-") | gsub("(^-|-$)"; "") | split("-")[0:6] | join("-"))"')
```

Then:

```bash
BRANCH_SAFE="${BRANCH//\//-}"
WORKTREE_PATH="${REPO_PATH}/../${REPO_NAME}.${BRANCH_SAFE}"
```

Default prompt (if none supplied):
```
Fix GitHub issue #<ISSUE> on branch <BRANCH>.
Run all tests before pushing.
Create a PR that closes #<ISSUE>.
```

## Phase 1 — Create worktree

```bash
wt -C "$REPO_PATH" switch --create "$BRANCH" --no-cd
```

Skip if `$WORKTREE_PATH` already exists.

## Phase 2 — Implement the fix

Work directly in `$WORKTREE_PATH`. Follow the prompt to implement, test, commit, and push. Create the PR:

```bash
gh pr create --title "..." --body "Closes #<ISSUE>" --head "$BRANCH"
```

## Phase 3 — Wait for PR

If the PR was created by a separate agent or process, poll every 60 seconds (up to 2 hours):

```bash
gh pr list --head "$BRANCH" --json number --jq '.[0].number'
```

`gh` auto-detects the repo from the git remote. Log progress every 10 minutes. If no PR after 2 hours, stop and report.

## Phase 4 — Wait for CI

```bash
gh pr checks "$PR_NUMBER" --watch --fail-fast
```

If CI fails: fix the issue in `$WORKTREE_PATH`, push, then re-run this phase.

## Phase 5 — Squash-merge

```bash
gh pr merge "$PR_NUMBER" --squash --delete-branch
```

## Phase 6 — Remove worktree

```bash
wt -C "$REPO_PATH" remove -D "$BRANCH"
```

## Parallel execution

Each issue runs in its own isolated worktree — no filesystem conflicts. Run multiple issues in parallel freely. If two branches modify the same file, GitHub will flag the merge conflict on the PR when it's time to merge; deal with it then.

**Integration test limitation:** Integration tests share a single Consul and Redis instance with hardcoded keys. Running `go test -tags=integration` from two worktrees simultaneously will cause key collisions and flaky failures. Only one worktree should run integration tests at a time — coordinate manually or stagger the test runs.
