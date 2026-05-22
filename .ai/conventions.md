# Development Conventions

Descriptive record of what the team has chosen. Unlike `rules/` which are normative, these describe actual practice.

## Branch Naming

- `feat/<feature-name>` for new features
- `fix/<bug-description>` for bug fixes
- `chore/<task>` for maintenance
- `main` is the trunk branch

## Commit Style

- Conventional commits: `type(scope): description`
- Types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`
- Scope: `auth`, `ws`, `game`, `ui`, `ai`, `admin`

## Code Review

- Reviewer checks: correctness, types, tests, scope, security
- PR merges to main preferred as squash

## Testing

- Go: standard `go test` with table-driven tests
- Frontend: Jest + React Testing Library
- Game logic must have tests for all hand types and play validation rules

## Environment

- `.env.example` documents all required variables
- Never commit `.env` files
- Frontend env vars prefixed with `NEXT_PUBLIC_`
