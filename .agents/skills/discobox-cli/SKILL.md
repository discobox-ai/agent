---
name: discobox-cli
description: Help the user use discobox through its CLI — answer how-to questions, explain what a command does, run the commands this box may run on their behalf, and hand them the exact command where only their own machine may. Use when the user asks how to do something with discobox (launch, attach, apply, ports, secrets, trust, harnesses, pools, servers, audit), when a discobox command fails for them, or when they ask what a box is doing.
---

# Using discobox for the user

You run inside a discobox. The CLI there talks to the user's server through
its pool, under the **sandbox role**. Some commands work from here; others
work only on the user's own machine. Always know which kind you are giving.

The in-image `discobox` skill describes this box itself: its source, its
ports, its meta. This skill is about the user's discobox as a whole.

## 1. Check the CLI, not your memory

The CLI changes between releases. Before you state a flag or a subcommand,
read `discobox <command> --help` (and `discobox admin <command> --help`) in
this box. Before relying on a behavior, read the design doc for it in
`discobox-ai/discobox`: the root `README.md` and `DESIGN.md`, `docs/`, and
`server/internal/resources/*/DESIGN.md`. When the user's version may differ,
ask for theirs (`discobox --version`).

## 2. Where a command can run

| Runs from this box (sandbox role) | Only on the user's machine |
|---|---|
| `new`, `ls`, `tag`, `rename`, `start`, `stop`, `restart` | `attach`, `shell`, `cp`, `proxy`, `apply`, `push` |
| `admin terminal screen\|input\|wait` on another box | `rm`, `admin box purge\|delete` |
| `secret request ls\|get\|approve\|deny` (within a delegation) | `secret create\|update`, `secret grant`, `secret rule` |
| `skills ls`, `admin audit` on boxes in the project | `configure`, `admin harnesses`, `admin pool`, `admin provider`, `admin remote`, `admin peer`, `uninstall` |

When a command is the user's to run:
- give it as one line they can paste;
- give the box by its name or ID;
- say what it will do and how to tell that it worked.

Never claim to have run it.

## 3. Common tasks

- **Launch work:**
  - `discobox new -d -C <dir|git-url>[@ref] -p "<prompt>"`.
  - `-H <harness>` picks the harness.
  - `--skills <dir|url[@ref][//path]>` adds skills.
  - `-d` creates the box and prints it without attaching.
- **Find a box:** `discobox ls`, filtered with `--tag k=v` or `--source`.
  Name it with `discobox rename` and label it with `discobox tag`.
- **Watch a box without attaching:**
  - `discobox admin terminal screen <box>` shows its screen.
  - `discobox admin terminal wait <box> --hook Stop` blocks until it goes
    idle.
  - `discobox admin audit list <box>` gives one timeline of its HTTP, DNS,
    credential verdicts, hooks and execs.
- **Bring changes home:** `discobox apply <box>` cherry-picks the box's
  commits onto the user's tree. `discobox push <box>` sends local commits the
  other way. Or the box opens a PR itself.
- **Reach a box's port:** `discobox proxy <box>`.
- **A box asks for a credential:** the user approves with
  `discobox secret request approve <id>`. `--auto` turns the approval into a
  standing rule.
- **A host's certificate is refused:** `discobox trust request ls`, then
  `approve`.
- **Something is wrong with the connection:** start from
  `discobox --version`, which prints both the client and server versions and
  so shows whether the server answers at all. Then `--debug`, which prints
  every API request.

## 4. Diagnosing a failure the user reports

1. Get the exact command, the full output, and `discobox --version`.
2. Reproduce it from here if the sandbox role allows. Otherwise read the code
   path in `discobox-ai/discobox` (`cli/internal/cli`).
3. If it is a bug, say so and offer to file it. Filing is delegated work
   like any other: a worker box running the triage skill reproduces and
   labels it.
