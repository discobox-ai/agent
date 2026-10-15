# Worker skills

These are the skills the manager gives every worker box, through
`discobox new --skills .agents/worker-skills`. They are not loaded into the
manager itself. The `setup` skill reads this catalog to offer tasks and to
check that the environment can do them.

- A **task** is what the manager hands a box: one task per box.
- A **step** is what a task uses on the way.
- `project` is read by every worker before any task. Setup writes it.

## Tasks

| Skill | What it does | What comes back | Requires |
|---|---|---|---|
| `triage-issue` | Reads one issue, reproduces it, writes a failing test, classifies and labels it | A labelled issue, a triage comment, and a verdict on whether it is ready to deliver | `github.read`, `github.triage`, `repo.runnable` |
| `deliver-issue` | Implements one issue and takes it to a pull request that passes CI, then keeps that PR mergeable | A green PR that waits for a human to merge it | `github.read`, `github.push`, `ci.read`, `repo.runnable` |
| `develop` | Builds a change already agreed with the user and delivers it as a PR | A green PR | `github.push`, `ci.read`, `repo.runnable` |

## Steps

| Skill | Used by | What it does |
|---|---|---|
| `commit` | every task that changes code | Groups changes into conventional commits |
| `verify` | `deliver-issue`, `develop` | Exercises the change at runtime and reports what was observed |
| `test-fix` | `deliver-issue`, `develop` | Runs the tests that cover a change, then a QA pass |
| `open-pr` | `deliver-issue`, `develop` | Pushes a branch, opens the PR, and fixes CI until it is green |
| `push-main` | a project that lands work by pushing straight to its default branch | Pushes, then fixes CI until it is green |

## Requirements

| Requirement | Met when |
|---|---|
| `github.read` | The box can read issues, PRs and code through `com.github.api` |
| `github.triage` | The box can comment on and label issues |
| `github.push` | The box can push branches and open PRs |
| `ci.read` | The box can read check runs and job logs for a commit |
| `repo.runnable` | The repository says how to build, test and run it, and a box can |

The skills were copied from `discobox-ai/skills` and are still written for
discobox's own development. Each has to be generalized to read the
repository's own docs and `project` before a non-discobox project can use it
(ADR 26-10-15-157 §11).
