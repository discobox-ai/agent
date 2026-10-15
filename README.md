# agent

A Claude Code agent that helps you use
[discobox](https://github.com/discobox-ai/discobox) and runs development
across the repositories you give it. It watches issues, pull requests and
notifications, tracks every piece of work, prioritizes it, delegates it to
worker discoboxes running the skills it keeps in `.agents/worker-skills`, and
brings you only the decisions that are yours. It records what the work
teaches it by editing those skills.

It is not a program you call. This repository *is* the agent: `AGENTS.md`
is its instructions and `.agents/skills` its own skills. Run it in a discobox
and talk to it there. It starts knowing nothing about your project: the
first conversation sets it up (which repositories, which tasks, how work
lands), and it records the answers in `agent.yaml`.

```sh
discobox new -C https://github.com/discobox-ai/agent@main -p "Set yourself up for my project."
```

The design is [ADR 26-10-15-157](docs/adr/26-10-15-157-the-manager-is-claude-code-woken-by-a-watcher.md).

## Build

```sh
go tool task build
```
