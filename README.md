# CPM (Construction Project Manager) — Go + React port

Rewrite of the original Laravel/Inertia Construction Project Manager app into a
decoupled architecture: **Go (Gin + GORM) REST API** backend and a **React +
TypeScript (Mantine UI)** frontend consuming it over `axios`/REST, with
realtime features over Pusher.

Original app (Laravel, kept as reference for behavior/UI parity) lives outside
this repo at `File_Skripsi_Priyatna_201011400103/Application`.

## Stack

- **Backend**: Go, Gin, GORM, PostgreSQL, Swagger (swaggo)
- **Frontend**: React, TypeScript, Mantine UI, axios
- **Realtime**: Pusher (Cloud)
- **Infra**: Docker Compose (Postgres), GitHub Actions CI/CD

## Repo layout

```
backend/    Go REST API (cmd/api, internal/{config,database,handlers,routes,models,middleware})
frontend/   React + TS SPA (Mantine UI)
docker-compose.yml   Local Postgres
```

## Running locally

```bash
# 1. Start Postgres
docker compose up -d

# 2. Backend
cd backend
cp .env.example .env
go run ./cmd/api
# API on :8080, Swagger UI on /swagger/index.html

# 3. Frontend (once scaffolded)
cd frontend
npm install
npm run dev
```

## Roadmap / module tracking

Tracked via GitHub Issues + Project board. See [Project board](../../projects)
and issues labeled per phase.

| Phase | Module | Status |
|---|---|---|
| 0 | Scaffold: Gin+GORM+Postgres+Docker, CI, Swagger, health check | ✅ done |
| 1 | Auth & RBAC (admin/manager/team member/client) | ⬜ todo |
| 2 | Owner/Client Company, Project, TaskGroup, Task CRUD + Labels | ⬜ todo |
| 3 | Work Report + EVM engine (cumulative-progress fix) | ⬜ todo |
| 4 | Realtime: Comments/discussion + EVM broadcast (Pusher) | ⬜ todo |
| 5 | Notifications: SMTP email alerts (EVM threshold) + activity log | ⬜ todo |
| 6 | Inventory & allocations | ⬜ todo |
| 7 | Reports/export (PDF/Excel) + Dashboard aggregation | ⬜ todo |

## Known issue carried over from the Laravel app (must fix, not port as-is)

The original app's EVM `progress` field is stored **incrementally per work
report** (frontend computes `work_done_this_report / task_volume`), but both
the task-metrics updater and the EVM service read it with `max('progress')`,
which assumes cumulative values. This under-counts earned value whenever a
task has more than one report. The Go port must store/derive **cumulative**
progress per task as of a given date, not `max()` over incremental values.

# Testing x account
Role	Email	Password
Manager	manager.test@example.com	manager123
Team Member	teammember.test@example.com	teammember123

