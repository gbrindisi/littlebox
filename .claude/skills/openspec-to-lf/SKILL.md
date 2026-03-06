---
name: openspec-to-lf
description: Convert an OpenSpec change's tasks.md into JSON tasks for littlefactory, with fat context, embedded todo checklist, labels, and blockers.
argument-hint: "<path-to-openspec OR path-to-change-dir OR path-to-tasks.md>"
disable-model-invocation: false
allowed-tools:
  - Read
  - Write
  - Grep
  - Glob
  - Bash(nl *)
  - Bash(sed *)
  - Bash(awk *)
  - Bash(rg *)
  - Bash(ls *)
  - Bash(cat *)
  - Bash(tree *)
---

# OpenSpec → Littlefactory JSON (Epics Only + Fat Context)

Create one JSON task per `##` section in `tasks.md`.
Do NOT create tasks for individual checklist items.

Output format: `.littlefactory/tasks.json`

## Hard Rules
- Do NOT invent references. Cite exact file+line ranges from proposal/design/specs.
- Generate unique task IDs using format: `<change-name>-<random-3char>`
- Tasks are ordered sequentially with blockers (task N+1 blocks on task N)
- Final JSON must be valid and well-formatted

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

Apply both labels to each task.

---

# Step 1 — Parse tasks.md into sections

Each section looks like:
## 1. Project Setup
- [ ] 1.1 ...
- [ ] 1.2 ...

Create exactly ONE task per section.
Inside the task description, include the todo list items verbatim (as a checklist).

---

# Step 2 — Sequencing with blockers

Sequential execution:
- Task 1: no blockers (status: "todo")
- Task 2: blockers = [task-1-id]
- Task 3: blockers = [task-2-id]
- etc.

This ensures tasks are worked on sequentially.

---

# Step 3 — Task description template (strict)

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

The task list can be found at <task.md path>

## Implementation plan (suggested chunks)
- Break the checklist into 3–8 chunky steps an agent can execute sequentially.
- Each step should mention concrete artifacts/files to produce.

## Acceptance criteria
- Bullets that are checkable (build/tests/commands, behaviors).
- "All checklist items above are implemented and verified."
- The task list at <task.md> is updated (checklist items implemented are marked as done)

## Key references
Citations:
- path/to/design.md:Lx-Ly (Heading: …)
- path/to/specs/.../spec.md:Lx-Ly (Heading: …)
- path/to/proposal.md:Lx-Ly (Heading: …)

If something can't be found:
- NOT FOUND: <what> (reason)
"""

## Citation method (must be exact)
- Search design.md first, then proposal.md, then spec.md files using `rg -n`.
- Use `nl -ba <file> | sed -n '<start>,<end>p'` to get line ranges.

Minimum: at least 1 citation if any upstream docs exist.

---

# Step 4 — Generate task IDs and build JSON structure

For each section:
1) Generate unique ID: `<change-name>-<random-3char>`
   - Use lowercase letters/numbers for random suffix
   - Example: `tui-output-abc`, `remove-bd-dependency-xyz`
2) Create task object with fields:
   - `id`: string (generated above)
   - `title`: string (section heading, without numeric prefix)
   - `description`: string (formatted as per template in Step 3)
   - `status`: string ("todo" for all tasks)
   - `labels`: array of strings (change label + section label)
   - `blockers`: array of strings (empty for first task, [previous-task-id] for others)

Maintain mapping:
- section_index → task_id

---

# Step 5 — Wire task blockers (strict)

For consecutive tasks N then N+1:
- Task N+1's `blockers` field = [task_N_id]

First task has empty blockers array: `[]`

---

# Step 6 — Write JSON file

1) Construct final JSON structure:
```json
{
  "tasks": [
    {
      "id": "change-abc",
      "title": "Section Title",
      "description": "...",
      "status": "todo",
      "labels": ["change-name", "section-label"],
      "blockers": []
    },
    {
      "id": "change-def",
      "title": "Section Title 2",
      "description": "...",
      "status": "todo",
      "labels": ["change-name", "section-label-2"],
      "blockers": ["change-abc"]
    }
  ]
}
```

2) Use Write tool to create `.littlefactory/tasks.json` at project root
3) Ensure JSON is properly formatted with 2-space indentation

---

# Final summary (print)
- Which change_root was selected
- Task IDs + titles
- Where JSON file was written
- Any missing references warnings
