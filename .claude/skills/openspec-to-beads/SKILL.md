---
name: openspec-to-beads
description: Convert an OpenSpec change’s tasks.md into Beads tasks, with fat context, embedded todo checklist, labels, deps, and sync.
argument-hint: "<path-to-openspec OR path-to-change-dir OR path-to-tasks.md>"
disable-model-invocation: false
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash(bd *)
  - Bash(nl *)
  - Bash(sed *)
  - Bash(awk *)
  - Bash(rg *)
  - Bash(ls *)
  - Bash(cat *)
  - Bash(tree *)
---

# OpenSpec → Beads (Epics Only + Fat Context)

Create one Beads epic per `##` section in `tasks.md`.
Do NOT create Beads tasks for individual checklist items.

## Hard Rules
- Do NOT use `bd edit`.
- Do NOT invent references. Cite exact file+line ranges from proposal/design/specs.
- Before saying “done/complete”, run: `bd sync --flush-only`.

---

# Step 0 — Resolve paths + labels

## Resolve change_root
- If `$ARGUMENTS` ends with `tasks.md`: change_root = parent
- Else if `$ARGUMENTS` is `openspec/changes/<change>/`: change_root = that
- Else: stop and ask for the right task.md file to parse.

Artifacts in change_root:
- tasks.md (required)
- proposal.md (preferred)
- design.md (preferred)
- specs/**/spec.md (preferred)

Labels:
- change label: `change-<change-dir-name>`
- section label: kebab-case section title (drop numeric prefix)

Apply both labels to each epic.

---

# Step 1 — Parse tasks.md into sections

Each section looks like:
## 1. Project Setup
- [ ] 1.1 ...
- [ ] 1.2 ...

Create exactly ONE epic per section.
Inside the epic description, include the todo list items verbatim (as a checklist).

---

# Step 2 — Priority + sequencing

Epic priority by section index:
1→0, 2→1, 3→2, 4→3, 5+→4

Strict sequencing:
- Epic N+1 depends on Epic N via `bd dep add`.

This should yield exactly one `bd ready` at the start.

---

# Step 3 — Epic description template (strict)

"""
## Context
- 2–5 sentences: what change this section belongs to and why (from proposal/design/specs).

## Scope
- 3–7 bullets summarizing what the section delivers.

## OpenSpec checklist (reference)
Copy verbatim from tasks.md:

- [ ] 1.1 ...
- [ ] 1.2 ...
...

The task list can be found at <task.md path >

## Implementation plan (suggested chunks)
- Break the checklist into 3–8 chunky steps an agent can execute sequentially.
- Each step should mention concrete artifacts/files to produce.

## Acceptance criteria
- Bullets that are checkable (build/tests/commands, behaviors).
- “All checklist items above are implemented and verified.”
- The task list at <task.md> is updated (checklist items implemented are marked as done)

## Key references
Citations:
- path/to/design.md:Lx-Ly (Heading: …)
- path/to/specs/.../spec.md:Lx-Ly (Heading: …)
- path/to/proposal.md:Lx-Ly (Heading: …)

If something can’t be found:
- NOT FOUND: <what> (reason)
"""

## Citation method (must be exact)
- Search design.md first, then proposal.md, then spec.md files using `rg -n`.
- Use `nl -ba <file> | sed -n '<start>,<end>p'` to get line ranges.

Minimum: at least 1 citation if any upstream docs exist.

---

# Step 4 — Create epics + label them

For each section:
1) `bd create --type=epic --priority=<p> --title="<section heading>" --description="<epic desc>" --json`
2) Extract ID:

Preferred (single-line JSON):
`... --json | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]\+\)".*/\1/p' | head -n1`

Fallback (multi-line JSON):
`... --json | rg -o '"id"\s*:\s*"[^"]+"' | head -n1 | sed 's/.*"id"\s*:\s*"\([^"]\+\)".*/\1/'`

3) `bd label add <epic_id> <change-label>`
4) `bd label add <epic_id> <section-label>`

Maintain mapping:
- section_index → epic_id

---

# Step 5 — Wire epic dependencies (strict)

For consecutive epics N then N+1:
- `bd dep add <epic_id(N+1)> <epic_id(N)>`

---

# Step 6 — Sanity checks + sync

1) `bd ready`
Expect: exactly one ready issue (Epic 1).

If more than one is ready, fix deps.

2) `bd sync --flush-only`

Only after this may you say “done/complete”.

---

# Final summary (print)
- Which change_root was selected
- Epic IDs + titles + priorities
- Any missing references warnings

