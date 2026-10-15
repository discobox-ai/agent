---
name: verify-recipes
description: How to build and drive discobox-agent to observe a change at its surfaces — the `init` CLI, the seeded manager directory, and a headless Claude Code session in that directory.
---

# Verify recipes

## Build

`go tool task build` writes `build/discobox-agent`, versioned by
`git describe` (`<sha>-dirty` on an uncommitted tree).

## `init`

Seed into a fresh `mktemp -d`, never into the checkout:

```bash
B=$PWD/build/discobox-agent; V=$(mktemp -d)
$B init $V/mgr
ls -la $V/mgr $V/mgr/.claude; cat $V/mgr/.discobox-agent/manifest.json
git -C $V/mgr log --oneline; git -C $V/mgr status --short
diff -r internal/templates/instance/.agents $V/mgr/.discobox-agent/base/.agents
```

Failure paths worth driving: a non-empty or file target, a missing nested
path, and no git identity, which makes the commit fail after the files are
written and so exercises cleanup:
`env HOME=$V/home GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 $B init $V/x/y/z`;
then `find $V` must show nothing `init` made.

## The seeded manager

`claude -p` run inside the seeded directory shows what the manager loads:
ask it for its instructions' first heading and its project skills. It shares
your user-level skills, so ask it to name only `.claude/skills` ones.
