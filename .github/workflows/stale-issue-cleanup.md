---
on:
  schedule:
    - cron: "0 5 * * 0"
  workflow_dispatch:
permissions:
  contents: read
  issues: read
safe-outputs:
  add-comment:
    title-prefix: "[stale]"
    max: 10
  add-labels:
    max: 10
  close-issue:
    max: 5
tools:
  - github[issues, labels]
  - bash
imports:
  - shared/mood.md
  - shared/issue-triage-go.md
engine: copilot
strict: true
timeout-minutes: 15
---

You are performing stale issue cleanup for this repository.

Goals:
- Scan open issues for staleness signals.
- Process warnings first, then closures.
- Limit actions to a maximum of 10 warnings and 5 closures per run.
- Be transparent: every action must include a comment explaining why.

Use GitHub issues and labels tools for issue inspection, comments, labels, and closures. Use bash only for date arithmetic and inactivity window calculations.

Staleness rules:
- No activity for 60+ days (including comments, label changes, and assignee changes) -> warning candidate.
- No activity for 90+ days after warning -> close candidate.

Warning phase:
- Add the `stale` label.
- Post a `[stale]` comment explaining that the issue has been inactive and will be closed in 30 days if there is still no activity.
- Invite the author or maintainers to respond if the issue is still relevant.

Close phase:
- Only close issues that were warned 30+ days ago and still have no activity.
- Post a `[stale]` closing comment explaining why the issue is being closed.
- Close no more than 5 issues in a single run.

Never close issues with any of the following:
- `P1-high`
- `security`
- `keep-open`
- Active assignees who have committed recently.
- Recent linked PR activity in the last 30 days.

Additional requirements:
- Do not surprise users with closures; warning must happen before closure.
- Warnings and closures must both include clear explanatory comments.
- Prioritize safety over throughput.
- If an issue has recent activity or any exemption applies, leave it open.
