# Claude Instructions

These instructions are for Claude-specific behavior in this repository. Generic project facts, architecture, setup, and API details live in `README.md`, `backend/README.md`, `frontend/README.md`, and `AGENTS.md`.

## Operating Mode

- Make the smallest safe change that solves the request.
- Read relevant files before editing.
- Prefer localized edits over broad refactors.
- Do not introduce new frameworks, dependencies, or deployment changes unless explicitly requested.
- For larger changes, state a short plan before editing.

## Response Style

- Be direct, technical, and concise.
- Prefer this shape for non-trivial work:
  - Plan: up to 3 bullets
  - Execution notes: only the important actions
  - Result: changed files and checks run
- Avoid long background explanations unless the user asks for them.

## Tool And Workflow Expectations

- Inspect the current implementation instead of relying on stale documentation.
- Preserve user changes in the working tree.
- Do not rewrite history, force-push, delete branches, or make destructive git changes without explicit instruction.
- Ask before pushing or opening a PR unless the user has clearly requested that workflow.

## Reporting

At the end of implementation work, report:

- What changed
- Which checks were run
- Which checks could not be run, if any
- Any follow-up risk that needs owner review
