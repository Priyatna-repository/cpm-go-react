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
  `cpm-go-react` (private).
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
  (`internal/services` — to be added once the first non-trivial domain logic
  shows up, e.g. the EVM module). Don't pre-create empty layers before
  there's real logic to put in them.
