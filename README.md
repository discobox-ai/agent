# agent

`discobox-agent` seeds a Claude Code agent that helps you use
[discobox](https://github.com/discobox-ai/discobox) and runs development
across the repositories you give it. The agent watches issues, pull requests
and notifications, tracks every piece of work, prioritizes it, delegates it
to worker discoboxes, and brings you only the decisions that are yours. It
records what the work teaches it by editing its own skills.

The agent is not a program you call: it is a directory of instructions and
skills that Claude Code runs. `discobox-agent init` writes that directory as
a new git repository. Run it in a long-lived discobox, start Claude Code in
the directory, and talk to it there. It starts knowing nothing about your
project: the first conversation sets it up (which repositories, which tasks,
how work lands), and it records the answers in `agent.yaml`.

```sh
go run github.com/discobox-ai/agent/cmd/discobox-agent@main init ~/manager
cd ~/manager && claude
```

The design is [ADR 26-10-15-157](docs/adr/26-10-15-157-the-manager-is-a-claude-code-agent-seeded-by-a-binary.md).

## Build

```sh
go tool task build
```
