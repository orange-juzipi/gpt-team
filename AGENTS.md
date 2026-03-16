# Repository Guidelines

## Repo Layout
- `frontend/`: Bun + Turbo monorepo for the admin UI.
- `frontend/apps/web`: Vite + React app.
- `frontend/packages/ui`: shared UI package.
- `go/`: Gin + GORM API service.

## Preferred Workflows

### Full-stack local development
- Start the API first:
  - `cd go && go run ./cmd/api`
- Start the frontend in a second terminal:
  - `cd frontend && bun run dev`
- Health check for the API:
  - `curl http://localhost:8088/api/health`

### Frontend
- Install dependencies from `frontend/` with `bun install`.
- Run all frontend tasks from `frontend/`:
  - `bun run dev`
  - `bun run test`
  - `bun run build`
  - `bun run typecheck`
  - `bun run lint`
- App-specific commands are available from `frontend/apps/web/`:
  - `bun run test:watch`
  - `bun run preview`
- To add shadcn/ui components, use:
  - `pnpm dlx shadcn@latest add <component> -c apps/web`

### Backend
- Run backend tasks from `go/`:
  - `go test ./...`
  - `go build ./cmd/api`
- Copy `go/.envrc.example` to `go/.envrc` for local setup. If using `direnv`, run `direnv allow`.

## Notes
- Frontend development expects `/api` to be proxied to the backend via `apps/web/.env.local`.
- Prefer the Bun/Turbo workspace commands from `frontend/` unless a task is clearly app-specific.
