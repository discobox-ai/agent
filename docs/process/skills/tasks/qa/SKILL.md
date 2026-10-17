---
name: qa
description: Verify one merged change on the default branch, in a fresh box once the merge commit's CI is green: build and run that commit, check it against the issue's acceptance criteria and repro (for a parent issue, the whole feature against its design), and post a QA report with evidence on the issue. Sets Verification to Passed for a human to sign off; or to Failed and reopens the issue as Ready for rework (Backlog if it has already failed once). Fixes nothing. Use when an issue's Verification is Pending and its merge commit's CI is green.
---

# QA

TODO
