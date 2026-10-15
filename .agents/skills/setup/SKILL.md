---
name: setup
description: Set this manager up, or change its setup, in conversation with the user — learn the environment, which tasks they want delegated, what this manager can do here and whether that matches, which repositories it manages and how work lands in each, and where its own changes are published; record it in agent.yaml and the `project` worker skill. Use on first start, when agent.yaml does not exist, or when the user asks to change what is managed, which tasks are delegated, how work is done, or where the manager's changes go.
---

# Setting up the manager

You arrive knowing nothing about this user's project. Nothing in this
repository says what you manage or how its work is done. Setup is how you
learn it, from the user and from the repositories themselves. You then
record it where you and your workers will read it.

ADR 26-10-15-157 §3 is the design.

## 0. Where you stand

- **No `agent.yaml` at the root of this checkout:** this is first setup. Say
  so in one line, then run §1–§8 in order.
- **`agent.yaml` exists:** the user wants something changed. Read the file
  and `.agents/worker-skills/project/SKILL.md`. Run only the sections the
  change touches, and show the user the diff before committing.

How to run the conversation:

- **One topic at a time,** in the order below. Each later step depends on
  the earlier ones.
- **Look first, then propose.** Before asking anything, find out what you
  can yourself. Then put what you found to the user to confirm or correct.
  Ask open questions only for what nothing could tell you.
- **Use `AskUserQuestion` for choices,** with your recommendation first.
- **Never fill a gap with a guess.** Record it as unknown, and say what
  stays unknown at the end.

## 1. The environment

Probe the environment, then summarize it in a short table for the user to
confirm:

- **Discobox.**
  - `discobox --version` for the client and server versions.
  - `discobox ls` to check that this box may list and create boxes.
  - `discobox skills ls` for the project's skills sources.
- **Credentials.** Run `discobox-access list` to see what this box may
  already use. Read the in-image `discobox-access` skill for what it can
  request.
- **Tools inside a box:** `docker info` (nested Docker), `gh --version`, and
  the languages the repositories need.
- **Where the code lives.** Ask which code host it is, unless the user has
  already named repositories. The watcher reads GitHub only. If the code is
  anywhere else, say that the manager can delegate work there but cannot
  watch it.

## 2. The tasks the user wants delegated

1. Read `.agents/worker-skills/README.md`, the catalog.
2. Show the user the **tasks**, one line each: what the task does, and what
   comes back to them when it is done.
3. Ask which tasks they want, as a multi-select.
4. If they want something the catalog does not have, record it as a wanted
   task. Offer to write its worker skill with them after setup. That is
   learning (`AGENTS.md`).

## 3. What this manager can do here

For each chosen task, check every requirement the catalog lists against what
§1 found. Give each requirement one of three answers:

- **met**;
- **can be requested,** for a credential this box can ask for;
- **missing,** with what would provide it.

## 4. Whether that matches

1. Show a table with these columns:
   - task;
   - whether it can be done here: yes, after a credential is requested, or
     no;
   - what is missing.
2. For every "no", let the user choose one of these:
   - drop the task;
   - provide what is missing;
   - adapt the task's skill to work without it.
3. Enable only tasks that can run.
4. Request the credentials the enabled tasks need, all in one round
   (`discobox-access request`).
   - Word each use so that a worker's later request matches its delegation
     word for word.
   - Record the wording, because the worker skills must use the same words.

## 5. The repositories

Get the list of repositories from the user, as `owner/name`. For each one:

1. **Read before asking.** Use `gh api`, run under `discobox-access`. Look
   at:
   - the default branch;
   - `AGENTS.md` or `CLAUDE.md`, `CONTRIBUTING*`, and `README`;
   - `.github/workflows` and the required checks on the default branch;
   - `.agents/skills`;
   - the labels in use.
2. **Propose how work lands there,** from what you read:
   - PR or direct push;
   - who merges;
   - which checks must pass;
   - branch naming;
   - commit style;
   - how issues are labelled and prioritized;
   - how a change is built, tested and run.
3. **Ask only about what the repository does not already say.** Where it
   documents something, record a pointer to it instead of a copy, because a
   worker cloned from the repository reads it there.

## 6. Publishing and limits

- **Where the manager's own commits go** (ADR §10):
  - `off`: commit locally only. This is the default.
  - `push`: push to a remote and branch.
  - `pr`: open a pull request there.

  For `push` or `pr`, point the user to a fork or a branch of their own. The
  template is shared, so this copy's setup must not land in it.
- **How many worker boxes may run at once.** Propose 4. Lower it if `df -h /`
  shows the host is short on disk.

## 7. Record it

1. Write `agent.yaml`:

   ```yaml
   # Written by the setup skill. To change it, ask the manager.
   environment:
     codeHost: github
     notes: one line per fact from §1 worth keeping
   tasks: [triage-issue, deliver-issue]   # enabled, from the catalog
   wantedTasks: []                         # asked for, not in the catalog yet
   repos:
     - name: owner/repo
       defaultBranch: main
   publish:
     mode: off                             # off | push | pr
     remote: ""                            # for push and pr
     branch: ""
   limits:
     maxBoxes: 4
   ```

2. Rewrite `.agents/worker-skills/project/SKILL.md` with what a worker needs
   and its repository does not say. Give each repository one section:
   - how work lands there (§5);
   - pointers to its own docs;
   - the credential uses, in the exact wording from §4.

   Keep the frontmatter as it is.
3. Show the user both files.
4. On their confirmation, commit with `chore(setup): <what changed>`.
5. Publish as §6 says. On first setup, ask before the first push.

## 8. Report

End with a short summary:

- the enabled tasks, and the repositories managed;
- what is still unknown or missing;
- what the user can ask for next, such as "triage the open issues in
  owner/repo".
