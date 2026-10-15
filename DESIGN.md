# agent design

This repository builds `discobox-agent`. The binary embeds the templates of
a development manager, a Claude Code agent, and seeds a manager from them.
The manager runs in a long-lived discobox, in a checkout that
`discobox-agent init` wrote. Its harness loads that checkout's `AGENTS.md`
(its role and rules) and `.agents/skills` (what it knows how to do), and acts
through the discobox CLI, `discobox-access` and `gh`. Work goes out to worker
discoboxes running the skills in its `.agents/worker-skills`.

```mermaid
flowchart LR
  bin[discobox-agent init] -- writes templates --> mgr
  user([User]) -- discobox attach --> mgr[Manager<br/>Claude Code + AGENTS.md]
  mgr -- discobox new --skills .agents/worker-skills --> w1[Worker box<br/>repo A]
  mgr -- discobox new --> w2[Worker box<br/>repo B]
  w1 & w2 -- tags: work, state, pr, needs --> mgr
  mgr -- gh via discobox-access --> gh[(GitHub org)]
  w1 & w2 -- PRs --> gh
```

- `internal/templates/instance`: every file a manager starts with, at the
  path it lands at, embedded with `all:` so the dot-directories come along.
  It ships without `agent.yaml`, which is how a new manager knows it is
  unconfigured. Its `CLAUDE.md` symlink is there for the repository
  standard and is never seeded as a template (see `internal/instance`).
  - `AGENTS.md`: the manager's instructions.
  - `.agents/skills/setup`: the onboarding conversation that writes
    `agent.yaml`.
  - `.agents/skills/discobox-cli`: helping the user use discobox: which
    commands run from the box and which only on their machine.
  - `.agents/worker-skills`: the skills each worker box is launched with
    (`discobox new --skills`).
- `internal/instance`: `Init` seeds a directory that is missing or empty. It
  writes the templates, the symlinks every instance has (`CLAUDE.md`,
  `.claude/skills`; a template at either path is skipped, because a checkout
  without symlinks hands embed a regular file there), and `.discobox-agent`: `manifest.json` (the seeding
  version and the template paths) and `base/`, a copy of each template as
  written. The base is what a later upgrade merges against, because the
  manager edits its files from then on. It then makes the directory a git
  repository with one commit. A failure removes what it wrote.
- `cmd/discobox-agent`: `init <dir>` and `--version`.
- `internal/version`: the build version, set by ldflags, falling back to the
  VCS revision, then to the module version Go records (`go run …@main`).
