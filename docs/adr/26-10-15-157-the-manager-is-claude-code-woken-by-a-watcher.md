# 26-10-15-157 — The manager is Claude Code woken by a watcher
- **Status**: Proposed
- **Date**: 2026-10-15

## Context

Development across the `discobox-ai` org runs in discoboxes: one box per
piece of work, driven by a harness and a skill. Coordinating those boxes was
the `orchestrate-issues` skill in `discobox-ai/discobox`. It was a chat
session acting as a lead, plus a shell loop (`orchestrate.sh watch`) that
scraped worker terminals for Claude Code's title glyph and strings such as
"esc to interrupt", and exited on every event so that an LLM turn had to
re-arm it. It covered one repository, kept its state in
`~/.local/state/orchestrate`, and only picked up issues when a person asked.

The user wants one agent that does two jobs:

- helps them **use discobox** through its CLI;
- **runs development** for the whole org: watches issues, pull requests and
  notifications, tracks the status of each piece of work, prioritizes it,
  delegates it to worker boxes, and brings to the user only what needs them.

The shared SDLC skills (triage, deliver, review, verify, open a PR) live in
`discobox-ai/skills`. Every box, for any repo in the org, pulls them in.

## Decision

1. **The manager is a Claude Code agent.** It is this repository: `AGENTS.md`
   is its instructions and `.agents/skills` holds its own skills. It runs as
   a long-lived discobox cloned from `discobox-ai/agent`, in the box's
   `primary` terminal, which is the Claude Code TUI. The manager coordinates
   and delegates; it does no implementation work itself.
   - **People** use that terminal directly: `discobox attach`.
   - **Programs** use it without attaching: `discobox admin terminal screen`
     reads it, `input` types into it, and `wait --hook Stop --after` follows
     its turns. Programs here means the watcher, the manager when it looks at
     a worker, and a later GUI.
   - **One mechanism for every box.** The same commands drive the manager and
     every worker, on any harness.

2. **A Go watcher observes, and wakes only the manager.**
   `discobox-agent watch` runs as a discobox service in the manager's box.
   - **What it polls:**
     - GitHub, for the configured org: issues, pull requests, check runs on
       PR heads and on `main`, and the user's notifications. It respects
       `X-Poll-Interval` and uses conditional requests.
     - Discobox: the boxes it launched, their tags, their `Stop` harness
       hooks, and pending secret requests.
   - **What it stores:** in SQLite (`github.com/discobox-ai/x/gormdb`):
     - what it observed;
     - the board, one row per work item;
     - an event queue.
   - **How it wakes the manager:** when events that need a decision are
     queued, it types a one-line notice into the manager's `primary` terminal
     with `terminal input`. The notice names the events and nothing else.
     - **Wait for a quiet terminal.** It types only when both hold:
       - the manager's last hook is `Stop`, so no turn is running;
       - the terminal's `lastAccessedAt` is no later than that `Stop`, so no
         one has typed since and the prompt is empty.

       Otherwise it holds the notice until the next `Stop`.
     - **Batch.** Events queued while a turn runs go out together in one
       notice.
     - Because the watcher drops everything that needs no decision, a notice
       always has work behind it.
   - **How the manager reads it:**
     - `discobox-agent events` prints the queued events as JSON, and
       `discobox-agent ack` acknowledges them.
     - `board` and `inbox` print the current state.
   - **The notice is its only action.** Apart from it, the watcher changes
     nothing outside its database. Every action (launch, message, approve,
     retire) is the manager's, so every action passes through Claude Code's
     permissions and the credential judge.

3. **Workers report through box tags, not their screen.** A worker is
   launched like this:

   ```
   discobox new -d -C https://github.com/discobox-ai/<repo>@main \
     --skills https://github.com/discobox-ai/skills -p "/<skill> <args>"
   ```

   It keeps these tags in `~/.discobox/meta.json`:

   | Tag | Value |
   |---|---|
   | `work` | the work item, e.g. `discobox-ai/review#12` |
   | `skill` | the skill it runs |
   | `state` | `working`, `waiting` (a question for a human), `blocked`, `ready` (PR green, awaiting merge), `done` |
   | `pr` | the PR number, once there is one |
   | `needs` | one line: what it is waiting for |

   The watcher reads these through the discobox API, and the `Stop` hook
   tells it when a box goes idle. The rewritten skills in
   `discobox-ai/skills` implement this contract.

4. **The board is ranked by fixed rules, which the manager may override with
   a stated reason.** Highest first:
   - red CI on a default branch, or a security report;
   - a worker waiting on a human;
   - a PR that is green and ready to merge;
   - priority labels;
   - age.

5. **Results go up in two places.**
   - **The manager's terminal:** a ranked status table and the inbox (the
     items whose next move is the user's).
   - **An Artifact dashboard,** whose shared database holds the board and
     inbox. A decision the user records there is read back by the manager.

   Merges stay the user's.

6. **The conversation is one client of the watcher's state.**
   - **What lives in the watcher's database:** the board, the inbox, the
     decisions the user records, and what the manager did about each item.
   - **How every client reads it:** through `discobox-agent` commands that
     print JSON. The TUI session, the dashboard, and any later GUI all read
     it the same way.
   - **What lives only in the transcript:** reasoning, never state.
   - **No transcript parsing.** Nothing reads Claude Code transcripts,
     because their format is internal and changes between releases. To read
     a turn's final text, use a `Stop` hook's `last_assistant_message`.

7. **Credentials are held by the manager's box.**
   - `ai.discobox.sandbox` (create, supervise, answer requests).
   - A read-only `com.github.api` use for the watcher. It takes a fresh value
     per poll cycle through the `agentcreds` client's `get`, naming its poll
     as the command.
   - A `delegate` grant that holds every use a worker may ask for. Within that
     delegation the manager approves a worker's request.

## Alternatives rejected

- **Keep `orchestrate-issues` and its shell loop.**
  - Screen scraping breaks whenever the harness changes its UI.
  - The loop needs an LLM turn after every event.
  - It covers only one repository.
- **A standalone Go service that calls the Claude API for judgment.** It
  would rebuild what a harness already provides (tools, permissions, skills,
  the user's terminal). It would also need an admin token outside the
  credential judge. The user wants a Claude Code agent.
- **A watcher that also acts**, for example auto-approving credentials or
  retiring merged boxes, as `orchestrate.sh` did. Keeping it read-only leaves
  one actor whose every action is visible and judged. The cost is one LLM
  turn per action.
- **Worker skills in this repository.** They are shared by every repo in the
  org and stay in `discobox-ai/skills`. This repository holds only what the
  manager runs.
- **Running the manager inside tmux** and driving it with `send-keys` and
  `capture-pane`.
  - tmux adds panes, captures with colors, and `pipe-pane`, but none of
    these is needed here.
  - tmux knows nothing about harness hooks, so the end of a turn would be
    guessed from the screen.
  - It reaches only the box it runs in, and nests one multiplexer inside
    another.
  - Workers already run in `primary`, so the manager would be driven
    differently from every worker.
- **Waking the manager with a background `discobox-agent next`** whose exit
  re-invokes Claude Code. It works only while the manager itself keeps the
  command running. It also wakes the session in the middle of what the user
  is typing.
- **Waking the TUI through a Claude Code channel.** A channel is an MCP
  server that pushes events into a live session. It is a research preview:
  it needs `--dangerously-load-development-channels` and auth through
  claude.ai or a Console API key. It also works only for Claude Code, while
  `terminal input` works on any harness.
- **Reading the conversation from transcript files** to rebuild state or to
  feed a GUI. The format is internal. State the GUI needs lives in the
  watcher (§6).
- **Box meta in a side database.** Tags already ride every box's status
  report, so the watcher reads them without touching the worker.

## Deferred

- **GitHub webhooks instead of polling.** A box has no public inbound
  endpoint. Revisit when discobox routes inbound webhooks to a box, or when
  polling hits rate limits at org scale.
- **A custom GUI.** Revisit when the TUI stops being enough.
  - The GUI hosts the conversation on the Claude Agent SDK. It resumes the
    TUI's session by ID, one at a time and never at the same moment, because
    nothing locks a transcript against two writers.
  - It reads the board and inbox from the watcher (§6).
  - It needs no change to the watcher, `AGENTS.md`, or the skills.

## Consequences

- Worker skills must keep the tag contract. A skill that does not keep it
  shows on the board only as idle or busy.
- The manager's box is long-lived. It holds a keepalive lease and its
  database survives restarts, so losing the box loses only state the watcher
  can rebuild from GitHub and discobox.
- One manager serves one user and one org. Several managers would need a
  shared board, which this does not provide.
