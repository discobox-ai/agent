# agent design

This repository is a Claude Code agent, not a service. The harness runs
in a discobox cloned from this repository. It loads `AGENTS.md` (the
manager's role and rules) and `.agents/skills` (what it knows how to do), and
acts through the discobox CLI, `discobox-access`, and `gh`. Work goes out
to worker discoboxes running the skills in `.agents/worker-skills`.

```mermaid
flowchart LR
  user([User]) -- discobox attach --> mgr[Manager<br/>Claude Code + AGENTS.md]
  mgr -- discobox new --skills .agents/worker-skills --> w1[Worker box<br/>repo A]
  mgr -- discobox new --> w2[Worker box<br/>repo B]
  w1 & w2 -- tags: work, state, pr, needs --> mgr
  mgr -- gh via discobox-access --> gh[(GitHub org)]
  w1 & w2 -- PRs --> gh
```

- `.agents/worker-skills`: the skills each worker box is launched with (`discobox new --skills`).
- `.agents/skills/discobox-cli`: helping the user use discobox: which
  commands run from the box and which only on their machine.
- `cmd/discobox-agent`: the agent's companion binary. It holds only its
  version for now.
- `internal/version`: the build version, set by ldflags, falling back to the
  VCS revision.
