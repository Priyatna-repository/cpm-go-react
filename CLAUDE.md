# CLAUDE.md

Context for Claude Code sessions working on this repo.

## What this is

A ground-up rewrite of a Laravel/Inertia construction-project-management app
into a decoupled Go REST API + React/TypeScript SPA. The **UI and behavior**
should match the original app (reference copy outside this repo, at
`../File_Skripsi_Priyatna_201011400103/Application`); the **backend
implementation** is a fresh, better-structured design, not a line-by-line
port. Frontend is being ported incrementally to TypeScript; it already uses
Mantine UI (kept as-is), so component styling largely carries over.

## Architecture decisions (settled, don't re-litigate without asking)

- Monorepo: `backend/` (Go) + `frontend/` (React/TS), single GitHub repo
  `cpm-go-react` (public — made public 2026-09-03 so branch protection could
  be enabled without a paid plan).
- Backend: Gin + GORM + PostgreSQL. Layout: `cmd/api` (entrypoint),
  `internal/{config,database,handlers,routes,models,middleware}`.
- API docs: swaggo/swag annotations on handlers, generated into `backend/docs`
  via `swag init -g cmd/api/main.go -o docs`, served at `/swagger/index.html`.
  Regenerate docs after adding/changing any handler's swag comments.
- Frontend talks to the backend over plain REST/JSON via axios — no Inertia.
  Client-side routing (React Router) replaces Inertia's server-driven routing.
- Realtime (task discussion comments, live EVM dashboard updates): Pusher
  Cloud, using the project's existing Pusher app credentials (see the
  reference Laravel app's `.env`). Go side publishes via
  `pusher/pusher-http-go`; frontend keeps using `pusher-js`.
- Progress tracking: GitHub Issues + a Project board, one issue per roadmap
  module (see README roadmap table). Commit messages should reference the
  issue they close (`closes #N`) so the board updates automatically.
- Auth (decided 2026-09-03, Phase 1): JWT access token (short-lived, ~15 min)
  + opaque refresh token (random value, only its SHA-256 hash stored in the
  `refresh_tokens` table, ~7 day TTL, delivered as an httpOnly cookie and
  rotated on every `/auth/refresh` call). Chosen over session-cookie auth
  because the SPA is a fully decoupled client. Roles are a simple `role_id`
  FK on `users` (admin/manager/team member/client, one role per user) rather
  than porting spatie/permission's granular per-action permission tables —
  Phase 1 only needs role-based middleware; revisit only if a later phase
  needs finer-grained permissions than "which role am I."

## Known correctness bug to fix, not carry over

The Laravel app's EVM `progress` per work report is **incremental**
(`work_done_this_report / task_volume`), but `max('progress')` is used
wherever cumulative task progress is needed (task metrics updater, EVM
service). This drastically under-counts earned value once a task has more
than one report. When designing the Go EVM module, progress must be derived
as a true cumulative value as of a given date — do not replicate `max()`
over incremental per-report values. See README's "Known issue" section for
the file:line references in the original app.

## Working conventions

- Every completed module/slice = its own commit(s) + closes its GitHub issue.
- Vertical slices: ship backend endpoint(s) + the frontend page consuming
  them together per module, not "all backend then all frontend" — the goal
  is an early demoable app, not a big-bang integration at the end.
- Keep handlers thin; business logic belongs in a service layer
  (`internal/services` — added in Phase 1 for `auth_service.go`, the first
  non-trivial domain logic: password hashing, JWT issuing, refresh-token
  rotation). Don't pre-create empty layers before there's real logic to put
  in them.

## How Claude should work in this repo

**Instructional mode, not autonomous execution.** For each unit of work,
give: (1) which file(s) are touched and why, (2) full file content or a
precise diff, (3) exact terminal commands to run — including `git
add/commit/push` — (4) which GitHub issue it closes, (5) how to verify it
worked. The user runs the commands himself; the point is that he understands
every change in a codebase he's shipping solo, not just review a finished
diff. Reading/investigating code, running builds/tests to verify what he
reports back, and small low-risk meta/doc edits (like this file) are fine to
do directly — the line is about who *writes and executes* application code
changes.

**Skills/agents available, and when they matter here:**
- `/code-review` (optionally `high`/`ultra`, or a PR number) — run on every
  PR's diff before merging. This is the practical stand-in for "does this
  change actually match the requirement," since that's a semantic judgment
  no CI check can make on its own.
- `/security-review` — run once Phase 1 (auth) and any handler touching
  credentials/tokens/permissions is in a PR.
- Everything else (Explore/general-purpose agents, TodoWrite) is Claude's own
  internal workflow during a session — not a gate on the repo, no setup
  needed.

**What actually rejects off-spec/broken changes automatically (GitHub-side,
not Claude-side):**
- `main` is branch-protected (enabled 2026-09-03): PRs required with 1
  approval (self-approval counts, stale approvals dismissed on new pushes),
  force-push and branch deletion blocked, and PRs cannot merge unless the
  `backend-ci.yml` `build-test` job (go vet/build/test, plus golangci-lint
  once added) is green. `enforce_admins` is left `false`, so the repo owner
  can still push directly to `main` in an emergency — that's a deliberate
  escape hatch, not a gap to close casually.
- Every phase issue (#2-#8 and beyond) carries acceptance criteria in its
  body — a PR should reference the issue (`closes #N`) and its description
  should map to those criteria, so review (`/code-review` or manual) has a
  concrete spec to check the diff against, not just "does it compile."
