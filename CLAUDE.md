# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Multiplayer **Poker(扑克)** card game platform with AI opponents, real-time WebSocket gameplay, and LiveKit voice/video. Starbucks-inspired design system. Admin panel for user/room/AI management.

## Engineering Principles

When facing ambiguity, prioritize in this order:

1. **Game correctness above all** — Card game logic must be bug-free. A broken game is worse than a slow UI. Validate all rule edge cases (bomb patterns, joker interactions, consecutive straights).
2. **Stability over novelty** — Prefer boring, proven patterns over clever abstractions. The existing architecture (`GameEngine` interface, WebSocket hub pattern, chi middleware chain) works — extend it, don't replace it.
3. **Type safety everywhere** — No `any` types in TypeScript. All Go interfaces must be explicit. The type system is your first line of defense against bugs.
4. **Simple over clever** — Three readable lines are better than one clever one-liner. Future AI sessions and human readers must understand the code without deep context.
5. **Respect existing patterns** — Look at how things are already done before making new things. New code should feel like it belongs: same error handling style, same file organization, same naming conventions.
6. **Test game logic thoroughly** — Game rule validation, hand evaluation, and play comparison must have tests. UI and wiring code can rely on type-checking.
7. **Consistent user experience** — UI changes must match the Starbucks-inspired design system. Don't introduce new visual patterns without checking `docs/DESIGN.md`.

## Architecture

**Server** (`server/`) — Go 1.26, chi router, gorilla/websocket, MySQL 8.0 via sqlx, JWT auth, bcrypt passwords

**Frontend** (`frontend/`) — Next.js 16, React 19, Tailwind CSS v4, TypeScript

### Game Flow

1. User logs in → gets JWT token
2. Joins/creates room via WebSocket (`join_room`)
3. When all players ready → game starts (`game_start`)
4. Players send actions (`room_action`), server broadcasts `state_update`
5. Round end → scores calculated, room resets to `waiting`

### Key Infrastructure

- **Docker Compose**: MySQL 8.0 + LiveKit server + Go server
- **Migrations**: `server/migrations/` — `001_users`, `002_game_tables`, `003_ai_tables`, `004_nickname_unique`
- **LiveKit**: WebRTC SFU server for voice/video (`livekit.yaml`)
- **Card encoding**: IDs 0–51 standard cards (suit×13 + rank), 52=small joker, 53=big joker. Rank: 3→3, 4→4, ..., A→14, 2→15, small joker→16, big joker→17.

## Commands

### Server

```bash
cd server

# Build for current platform
go build -o server ./cmd/server/

# Build cross-platform
./build.sh linux amd64

# Run locally (requires MySQL on localhost:3306)
go run ./cmd/server/

# Run with docker-compose (MySQL + LiveKit + server)
docker compose up --build

# Run tests in a specific package
go test ./internal/game/doudizhu/...
go test ./internal/api/...
go test ./internal/model/...

# Run a single test
go test ./internal/game/doudizhu/ -run TestHandType

# Run with verbose output
go test -v ./internal/api/...

# Manual server start
./run.sh
```

### Frontend

```bash
cd frontend

# Development server (default :3000)
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint
npm run lint

# Type check
npx tsc --noEmit

# Run tests
npm test
```

### Database

Migrations run automatically via Docker Compose volume mount into `mysql` container. For local dev, apply manually:

```bash
mysql -u root -p pokeface < server/migrations/001_create_users.sql
```

### Environment

See `.env.example` for required variables:
- `DATABASE_DSN` — MySQL connection string
- `JWT_SECRET` — JWT signing key
- `ALLOWED_ORIGINS` — CORS origins (comma-separated)
- `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET` / `LIVEKIT_HOST` — LiveKit credentials
- Frontend uses `NEXT_PUBLIC_API_URL` (default: `http://localhost:8080`) and `NEXT_PUBLIC_WS_URL`

---

## Workflow Rules (MANDATORY)

1. **Plan first, execute second** — Before touching any code, state the plan (files to modify, approach). Wait for confirmation.
2. **One task at a time** — Do not proceed to the next task until the current one is verified complete.
3. **Report after each step** — After each logical step, report what was done and the result (success / failure / diff).
4. **No scope creep** — Stick to the stated task. If you discover related issues, flag them but do not fix unless explicitly instructed.
5. **Ask when unsure** — If a requirement is ambiguous, ask before proceeding.

## Context Rules (MANDATORY)

Before starting work, read the relevant context files for the task type:

| Task Type | Must Read Before Starting |
|---|---|
| **Game logic** (rules, hand eval, play validation) | `docs/architecture/overview.md`, `docs/decisions/README.md` (ADR-002), `internal/game/doudizhu/` |
| **API / WebSocket** (handlers, hub, message routing) | `docs/architecture/api-rules.md`, `docs/architecture/overview.md` |
| **Frontend UI** (pages, components) | `docs/DESIGN.md`, `rules/frontend.md`, `docs/architecture/overview.md` |
| **AI / LLM** (AI players, prompts, provider) | `docs/decisions/README.md` (ADR-003), `internal/ai/` |
| **Auth / Security** | `rules/security.md`, `docs/architecture/api-rules.md` |
| **Admin panel** | `docs/architecture/api-rules.md`, `rules/frontend.md` |
| **Database / Model** (stores, queries, migrations) | `docs/decisions/README.md` (ADR-001), `docs/architecture/overview.md` |

**When you learn something new about the codebase** — a hidden constraint, a non-obvious pattern, a recurring bug — record it in `.ai/memory.md` under the appropriate section.

**When you make an architectural decision** — write an ADR in `docs/decisions/` following the existing format.

## Coding Rules

- **No `any` types** — Use proper TypeScript types / Go interfaces everywhere.
- **No mass refactoring** — Refactor only the code directly relevant to the task at hand. Do not restructure unrelated code.
- **Server language** — Go only. No other languages in `server/`.
- **Frontend language** — TypeScript + Tailwind CSS v4 only. No plain JavaScript in `frontend/`.
- **No new dependencies** — Do not add new npm / Go modules without explicit approval.
- **Keep it simple** — Prefer straightforward solutions over clever abstractions.

## Verification (MANDATORY)

After any code change, run ALL applicable checks and show their output verbatim:

- **Server changes**: `go vet ./...` + `go test ./...` (or the relevant package(s))
- **Frontend changes**: `npm run lint` + `npx tsc --noEmit` + `npm test`
- **Both**: If both sides changed, run both sets of checks.

Do not claim success — show the actual output. If checks fail, fix before proceeding.

## Harness Structure

Additional project infrastructure beyond source code:

| Directory | Purpose |
|---|---|
| `rules/` | Domain constraints — coding standards, frontend rules, security requirements |
| `scripts/` | Automated verification gates — `verify.sh`, `ci-local.sh` |
| `prompts/` | Reusable AI prompt templates — code review, refactoring, debugging |
| `docs/` | Long-term memory — architecture (overview, API rules), ADRs, business docs, known issues |
| `tasks/` | Current work tracking — backlog, doing, review, done |
| `.ai/` | Cross-session memory — persistent decisions, conventions, session logs |

## Scope Constraints

**Allowed to modify:**
- `server/internal/` — any Go source (handlers, middleware, stores, game logic, AI)
- `frontend/` — any TypeScript/TSX/Tailwind source (pages, components, lib)
- Migrations that are already applied (only additive changes)
- `CLAUDE.md`, `rules/`, `scripts/`, `prompts/`, `docs/`, `tasks/`, `.ai/`

**Forbidden to modify without explicit instruction:**
- `server/migrations/` — creating new migration files (requires separate review)
- `server/cmd/server/` — main entry point changes
- `docker-compose.yml` or Dockerfile
- `livekit.yaml` or any LiveKit configuration
- CI/CD configuration (GitHub Actions, etc.)
- `server/go.mod` / `frontend/package.json` — dependency changes
- `.env.example` or environment variable schema changes
- Any build/deploy scripts (`build.sh`, `run.sh`)

## Task Granularity

- **One logical change per task** — A "task" is a single feature, bug fix, or refactoring.
- **Max 3 files per step** — If a change touches more than 3 files, split it into sub-steps.
- **Conventional commits** — Commit messages follow conventional commits format: `type(scope): description` (e.g. `feat(auth): add guest login`, `fix(ws): handle reconnect race`, `refactor(game): extract hand evaluator`).
