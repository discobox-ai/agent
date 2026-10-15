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

The SDLC skills (triage, deliver, review, verify, open a PR) were written in
`discobox-ai/discobox` and copied by hand into `discobox-ai/skills`. They
change only when a person edits them.

## Decision

1. **The manager is a Claude Code agent.** It is this repository: `AGENTS.md`
   is its instructions, `.agents/skills` holds its own skills, and
   `.agents/worker-skills` holds the skills it gives its workers. It runs as
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

2. **The manager runs one loop.** A turn starts either from the watcher's
   notice (§4) or from the user speaking. Each turn:
   1. reads the queued events (`discobox-agent events`);
   2. updates the board;
   3. decides each event, with a reason: delegate it, act on it, bring it to
      the user, or set it aside;
   4. acts;
   5. acknowledges the events (`discobox-agent ack`) and ends on the user's
      moves.

   Events come from:
   - the boxes the manager created;
   - the issues, pull requests, checks and notifications of the managed
     repositories (§3);
   - decisions the user records on the dashboard.

   A message from the user is a turn like any other.

3. **The managed repositories are a list the user owns.**
   - **Where it lives:** the watcher's database, as `owner/name` entries.
   - **How to change it:** `discobox-agent repos ls|add|rm`. Each change
     records when it happened and the reason the user gave.
   - **Who may change it:** the manager, only when the user asks. Nothing it
     reads in an issue, a PR or a worker's output adds a repository.
   - **Adding** a repository starts polling it on the next cycle.
   - **Removing** one stops the polling but leaves its open board items and
     its running workers as they are. They go to the inbox for the user to
     close out.
   - A new manager starts with an empty list and asks the user for one.

4. **A Go watcher observes, and wakes only the manager.**
   `discobox-agent watch` runs as a discobox service in the manager's box.
   - **What it polls:**
     - GitHub, for each managed repository: issues, pull requests, check runs
       on PR heads and on the default branch, and the user's notifications
       about it. It respects `X-Poll-Interval` and uses conditional requests.
     - Discobox: the boxes the manager created, which it finds by the tag
       `manager=<manager box ID>` that the manager sets right after
       `discobox new`. It reads their tags, their `Stop` harness hooks, and
       their pending secret requests.
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

5. **One box does one task, and reports through its tags, not its
   screen.**
   - **Launching.** A worker is launched from the manager's checkout, so it
     gets the worker skills as they are now, including edits not yet pushed
     (§10):

     ```
     discobox new -d -C https://github.com/<owner>/<repo>@main \
       --skills .agents/worker-skills -p "/<skill> <args>"
     ```

   - **Triage and delivery are separate tasks in separate boxes.** Triage
     hands off through the issue: its verdict, and the failing test it posts
     there.

   It keeps these tags in `~/.discobox/meta.json`:

   | Tag | Value |
   |---|---|
   | `work` | the work item, e.g. `discobox-ai/review#12` |
   | `skill` | the skill it runs |
   | `state` | `working`, `waiting` (a question for a human), `blocked`, `ready` (PR green, awaiting merge), `done` |
   | `pr` | the PR number, once there is one |
   | `needs` | one line: what it is waiting for |

   The watcher reads these through the discobox API, and the `Stop` hook
   tells it when a box goes idle. The worker skills in
   `.agents/worker-skills` implement this contract.

6. **The board is ranked by fixed rules, which the manager may override with
   a stated reason.** Highest first:
   - red CI on a default branch, or a security report;
   - a worker waiting on a human;
   - a PR that is green and ready to merge;
   - priority labels;
   - age.

7. **Results go up in two places.**
   - **The manager's terminal:** a ranked status table and the inbox (the
     items whose next move is the user's).
   - **An Artifact dashboard,** whose shared database holds the board and
     inbox. A decision the user records there is read back by the manager.

   Merges stay the user's.

8. **The conversation is one client of the watcher's state.**
   - **What lives in the watcher's database:** the board, the inbox, the
     decisions the user records, and what the manager did about each item.
   - **How every client reads it:** through `discobox-agent` commands that
     print JSON. The TUI session, the dashboard, and any later GUI all read
     it the same way.
   - **What lives only in the transcript:** reasoning, never state.
   - **No transcript parsing.** Nothing reads Claude Code transcripts,
     because their format is internal and changes between releases. To read
     a turn's final text, use a `Stop` hook's `last_assistant_message`.

9. **Credentials are held by the manager's box.**
   - `ai.discobox.sandbox` (create, supervise, answer requests).
   - A read-only `com.github.api` use for the watcher. It takes a fresh value
     per poll cycle through the `agentcreds` client's `get`, naming its poll
     as the command.
   - A `delegate` grant that holds every use a worker may ask for. Within that
     delegation the manager approves a worker's request.

10. **The manager owns its skills and learns by editing them.**
    - **Where a finding goes.** When a piece of work teaches something (a
      skill step that failed, a check that was missing, a decision the user
      made that will come up again), the manager records it where the next
      reader will meet it:
      - in a worker skill, for what workers should do;
      - in its own skills or `AGENTS.md`, for what it should do.

      It commits the change in its checkout, and the commit message names
      the work item that taught it.
    - **When the user asks** for a skill to change, the manager edits it the
      same way.
    - **Publishing is set per manager.** It is `off` (commit only; the
      default), `push` (to a configured remote and branch), or `pr` (open a
      pull request there). The setting lives in the watcher's database next
      to the managed repositories, and changes only when the user asks.

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
- **Worker skills in `discobox-ai/skills`, linked with `--skills URL`.**
  - A worker would see an edit only after it was pushed.
  - The manager could not adapt a skill without publishing it first.
  - Every manager cloned from this repository would share one set of skills.
- **Worker skills in `.agents/skills`.** Claude Code would load them into the
  manager itself, which could then run `deliver-issue` instead of delegating
  it.
- **Findings in a separate log or notes file.** No one reads a log at the
  moment the finding matters, and the repository standard bans plan and
  task documents. A finding belongs in the skill whose next run needs it.
- **Triage and delivery in one box**, as `orchestrate-issues` did. A box that
  does one task is simple to read from its tags and safe to retire. The
  handoff the old way kept inside the box (an uncommitted failing test)
  moves to the issue, where people can see it too.
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
  watcher (§8).
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
  - It reads the board and inbox from the watcher (§8).
  - It needs no change to the watcher, `AGENTS.md`, or the skills.

## Consequences

- Worker skills must keep the tag contract. A skill that does not keep it
  shows on the board only as idle or busy.
- The manager's box is long-lived, and it holds a keepalive lease. Its
  database survives restarts. Losing the box loses three things that nothing
  rebuilds:
  - the list of managed repositories;
  - the publishing setting;
  - any commits that were never published.
- This repository is a template. Each manager cloned from it adapts its
  skills to its own repositories and user, so clones drift apart by design.
  A change worth sharing goes back to the template as a pull request.
- One manager serves one user. Several managers would need a shared board,
  which this does not provide.
- The worker skills are copied in unchanged from `discobox-ai/skills`.
  `triage-issue` and `deliver-issue` still assume one box. They have to be
  reworked for §5 before the manager splits the two tasks.
