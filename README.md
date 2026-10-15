# agent

The `discobox-ai` manager: a Claude Code agent that helps you use
[discobox](https://github.com/discobox-ai/discobox) and runs development
across the `discobox-ai` org. It watches issues, pull requests and
notifications, tracks every piece of work, prioritizes it, delegates it to
worker discoboxes running the shared
[skills](https://github.com/discobox-ai/skills), and brings you only the
decisions that are yours.

It is not a program you call. This repository *is* the agent: `AGENTS.md`
is its instructions and `.agents/skills` its skills. Run it in a discobox
and talk to it there:

```sh
discobox new -C https://github.com/discobox-ai/agent@main -p "What needs me today?"
```

The design is [ADR 26-10-15-157](docs/adr/26-10-15-157-the-manager-is-claude-code-woken-by-a-watcher.md).

## Build

```sh
go tool task build
```
